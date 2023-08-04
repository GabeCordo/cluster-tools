package core

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/module"
	"github.com/GabeCordo/etl/framework/components/provisioner"
	"github.com/GabeCordo/etl/framework/components/supervisor"
	"github.com/GabeCordo/etl/framework/utils"
	"math/rand"
	"reflect"
	"time"
)

const (
	DefaultHardTerminateTime = 30 // minutes
)

var provisionerInstance *provisioner.Provisioner

func GetProvisionerInstance() *provisioner.Provisioner {

	if provisionerInstance == nil {
		provisionerInstance = provisioner.NewProvisioner()
	}
	return provisionerInstance
}

func (provisionerThread *ProvisionerThread) Setup() {

	provisionerThread.accepting = true

	// initialize a provisioner instance with a common module
	GetProvisionerInstance()
}

func (provisionerThread *ProvisionerThread) Start() {

	provisionerThread.listenersWg.Add(1)

	go func() {
		// request coming from http_server
		for request := range provisionerThread.C5 {
			if !provisionerThread.accepting {
				fmt.Println("closing C5")
				provisionerThread.listenersWg.Done()
				break
			}
			provisionerThread.requestWg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessIncomingRequests(&request)
		}

		provisionerThread.requestWg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C8 {
			if !provisionerThread.accepting {
				fmt.Println("closing C8")
				break
			}

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessesIncomingDatabaseResponses(response)
		}

		provisionerThread.requestWg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C10 {
			if !provisionerThread.accepting {
				fmt.Println("closing C10")
				break
			}

			// if this doesn't spawn its own thread we can be left waiting
			provisionerThread.ProcessIncomingCacheResponses(response)
		}

		provisionerThread.requestWg.Wait()
	}()

	for _, moduleWrapper := range GetProvisionerInstance().GetModules() {

		for _, clusterWrapper := range moduleWrapper.GetClusters() {

			if clusterWrapper.IsMounted() && clusterWrapper.IsStream() {

				if !moduleWrapper.IsMounted() {
					moduleWrapper.Mount()
				}

				provisionerThread.logger.Printf("%s[%s]%s mount cluster \n", utils.Green, clusterWrapper.Identifier, utils.Reset)

				provisionerThread.C5 <- threads.ProvisionerRequest{
					Action:      threads.ProvisionerProvision,
					Source:      threads.Provisioner,
					ModuleName:  moduleWrapper.Identifier,
					ClusterName: clusterWrapper.Identifier,
					Metadata:    threads.ProvisionerMetadata{},
					Nonce:       rand.Uint32(),
				}
			}
		}
	}

	provisionerThread.listenersWg.Wait()
	provisionerThread.requestWg.Wait()
}

func (provisionerThread *ProvisionerThread) ProcessIncomingRequests(request *threads.ProvisionerRequest) {

	if request.Action == threads.ProvisionerMount {
		provisionerThread.ProcessMountRequest(request)
	} else if request.Action == threads.ProvisionerUnMount {
		provisionerThread.ProcessUnMountRequest(request)
	} else if request.Action == threads.ProvisionerProvision {
		provisionerThread.ProcessProvisionRequest(request)
	} else if request.Action == threads.ProvisionerTeardown {
		// TODO - not implemented
	} else if request.Action == threads.ProvisionerLowerPing {
		provisionerThread.ProcessPingProvisionerRequest(request)
	} else if request.Action == threads.ProvisionerModuleLoad {
		provisionerThread.ProcessAddModule(request)
	} else if request.Action == threads.ProvisionerModuleDelete {
		provisionerThread.ProcessDeleteModule(request)
	}
}

func (provisionerThread *ProvisionerThread) Send(request *threads.ProvisionerRequest, response *threads.ProvisionerResponse) {

	if request.Source == threads.Http {
		fmt.Println("sending provisioner response")
		provisionerThread.C6 <- *response
	}
}

func (provisionerThread *ProvisionerThread) ProcessAddModule(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerThread.logger.Printf("registering module at %s\n", request.Metadata.ModulePath)

	remoteModule, err := module.NewRemoteModule(request.Metadata.ModulePath)
	if err != nil {
		provisionerThread.logger.Alertln("cannot find remote module")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "cannot find remote module"}
		provisionerThread.Send(request, response)
		return
	}

	moduleInstance, err := remoteModule.Get()
	if err != nil {
		provisionerThread.logger.Alertln("module built with older version")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "module built with older version"}
		provisionerThread.Send(request, response)
		return
	}

	if err := GetProvisionerInstance().AddModule(moduleInstance); err != nil {
		provisionerThread.logger.Alertln("a module with that identifier already exists or is corrupt")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "a module with that identifier already exists or is corrupt"}
		provisionerThread.Send(request, response)
		return
	}

	// note: the module wrapper should already be defined so there is no need to validate
	moduleWrapper, _ := GetProvisionerInstance().GetModule(moduleInstance.Config.Identifier)

	registeredClusters := moduleWrapper.GetClusters()

	// REGISTER ANY HELPERS TO CLUSTERS THAT HAVE DEFINED THEM WITHIN THE STRUCT
	for _, clusterWrapper := range registeredClusters {

		clusterImplementation := clusterWrapper.GetClusterImplementation()

		reflectedClusterImpl := reflect.ValueOf(clusterImplementation)

		helperField := reflect.Indirect(reflectedClusterImpl).FieldByName("Helper")
		if helperField.IsValid() && helperField.CanSet() {
			if helper, err := NewHelper(provisionerThread.C9, provisionerThread.C11); err == nil {
				helperField.Set(reflect.ValueOf(helper))
			}
		}
	}

	// get all the configs from the original list that were sucessfully created
	registeredExports := make([]module.Cluster, 0)
	for _, export := range moduleInstance.Config.Exports {
		if _, found := moduleWrapper.GetCluster(export.Cluster); found {
			registeredExports = append(registeredExports, export)
		}
	}

	registeredConfigs := make(map[string]cluster.Config)

	// REGISTER EACH CONFIG FROM THE MODULE FILE TO THE DATABASE THREAD
	for _, export := range registeredExports {

		config := cluster.Config{
			Identifier:                  export.Cluster,
			OnCrash:                     export.Config.OnCrash,
			OnLoad:                      export.Config.OnLoad,
			StartWithNLoadClusters:      export.Config.Static.LFunctions,
			StartWithNTransformClusters: export.Config.Static.TFunctions,
			ETChannelThreshold:          export.Config.Dynamic.TFunction.Threshold,
			ETChannelGrowthFactor:       export.Config.Dynamic.TFunction.GrowthFactor,
			TLChannelThreshold:          export.Config.Dynamic.LFunction.Threshold,
			TLChannelGrowthFactor:       export.Config.Dynamic.LFunction.GrowthFactor,
		}

		if !config.Valid() {
			config = cluster.DefaultConfig
			config.Identifier = export.Cluster
		}

		registeredConfigs[config.Identifier] = config

		databaseRequest := threads.DatabaseRequest{
			Action:  threads.DatabaseStore,
			Nonce:   rand.Uint32(),
			Origin:  threads.Provisioner,
			Type:    threads.ClusterConfig,
			Module:  moduleInstance.Config.Identifier,
			Cluster: export.Cluster,
			Data:    config,
		}
		provisionerThread.C7 <- databaseRequest

		timeout := false
		var response threads.DatabaseResponse

		timestamp := time.Now()
		for {
			if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
				timeout = true
				break
			}

			if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(request.Nonce); found {
				response = (responseEntry).(threads.DatabaseResponse)
				break
			}
		}

		if timeout || !response.Success {
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "could not save config"}
			provisionerThread.Send(request, response)
			return
		}
	}

	provisionerThread.logger.Printf("dynamically loaded module %s\n", moduleInstance.Config.Identifier)

	// FOR EVERY CLUSTER WITH A MODE OF 'STREAM' PROVISION IT
	for _, export := range registeredExports {

		if !export.StaticMount {
			continue
		}

		provisionRequest := threads.ProvisionerRequest{
			Action:      threads.ProvisionerProvision,
			ModuleName:  moduleInstance.Config.Identifier,
			ClusterName: export.Cluster,
			Metadata: threads.ProvisionerMetadata{
				ConfigName: export.Cluster,
				Other:      make(map[string]string),
			},
			Source: threads.Provisioner,
			Nonce:  rand.Uint32(),
		}

		provisionerThread.C5 <- provisionRequest
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce, Description: "module registered"}
	provisionerThread.Send(request, response)
}

func (provisionerThread *ProvisionerThread) ProcessDeleteModule(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerInstance := GetProvisionerInstance()

	response := &threads.ProvisionerResponse{Nonce: request.Nonce}
	if deleted, _, found := provisionerInstance.DeleteModule(request.ModuleName); found {
		response.Success = true
		if deleted {
			response.Description = "module deleted"

			databaseRequest := threads.DatabaseRequest{
				Action: threads.DatabaseDelete,
				Type:   threads.ClusterModule,
				Module: request.ModuleName,
				Nonce:  rand.Uint32(),
			}
			provisionerThread.C7 <- databaseRequest

			databasePingTimeout := false
			var databaseResponse threads.DatabaseResponse

			timestamp := time.Now()
			for {
				if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
					databasePingTimeout = true
					break
				}

				if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
					databaseResponse = (responseEntry).(threads.DatabaseResponse)
					break
				}
			}

			if databasePingTimeout || !databaseResponse.Success {
				response.Success = false
				response.Description = "could not delete clusters and statistics under a module"
			}
		} else {
			response.Description = "module marked for deletion, a cluster is likely running right now, try later"
		}
	} else {
		response.Success = false
		response.Description = "module not found"
	}

	provisionerThread.Send(request, response)
}

func (provisionerThread *ProvisionerThread) ProcessPingProvisionerRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	if GetConfigInstance().Debug {
		provisionerThread.logger.Println("received ping over C5")
	}

	databaseRequest := threads.DatabaseRequest{
		Action: threads.DatabaseLowerPing,
		Nonce:  rand.Uint32(),
	}
	provisionerThread.C7 <- databaseRequest

	databasePingTimeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			databasePingTimeout = true
			break
		}

		if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	if databasePingTimeout || !databaseResponse.Success {
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		return
	}

	if GetConfigInstance().Debug {
		provisionerThread.logger.Println("received ping over C8")
	}

	cacheRequest := threads.CacheRequest{Action: threads.CacheLowerPing, Nonce: rand.Uint32()}
	provisionerThread.C9 <- cacheRequest

	cachePingTimeout := false
	var cacheResponse threads.CacheResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			cachePingTimeout = true
			break
		}

		if response, found := provisionerThread.cacheResponseTable.Lookup(cacheRequest.Nonce); found {
			cacheResponse = (response).(threads.CacheResponse)
			break
		}
	}

	if cachePingTimeout || !cacheResponse.Success {
		provisionerThread.logger.Alertln("failed to receive ping over C10")
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		return
	}

	if GetConfigInstance().Debug {
		provisionerThread.logger.Println("[etl_provisioner] received ping over C10")
	}

	response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: true}
	provisionerThread.Send(request, response)
}

func (provisionerThread *ProvisionerThread) ProcessMountRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.Send(request, response)
		return
	}

	if !moduleWrapper.IsMounted() {
		moduleWrapper.Mount()

		if GetConfigInstance().Debug {
			provisionerThread.logger.Printf("%s[%s]%s Mounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			provisionerThread.logger.Printf("%s[%s]%s could not find cluster\n", utils.Green, request.ModuleName, utils.Reset)
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.Send(request, response)
			return
		}

		if !clusterWrapper.IsMounted() {
			clusterWrapper.Mount()

			if GetConfigInstance().Debug {
				provisionerThread.logger.Printf("%s[%s]%s Mounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}

			if clusterWrapper.IsStream() {
				fmt.Println("sending provision request from provisioner")
				provisionerThread.C5 <- threads.ProvisionerRequest{
					Action:      threads.ProvisionerProvision,
					Source:      threads.Provisioner,
					ModuleName:  request.ModuleName,
					ClusterName: request.ClusterName,
					Metadata:    threads.ProvisionerMetadata{},
					Nonce:       rand.Uint32(),
				}
				fmt.Println("sent provision request to provisioner")
			}
		}
	}

	fmt.Println("sending response to provisioner")

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.Send(request, response)
}

func (provisionerThread *ProvisionerThread) ProcessUnMountRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.Send(request, response)
		return
	}

	if moduleWrapper.IsMounted() {
		moduleWrapper.UnMount()

		if GetConfigInstance().Debug {
			provisionerThread.logger.Printf("%s[%s]%s UnMounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.Send(request, response)
			return
		}

		if clusterWrapper.IsMounted() {
			clusterWrapper.UnMount()

			if GetConfigInstance().Debug {
				provisionerThread.logger.Printf("%s[%s]%s UnMounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}

			if clusterWrapper.Mode == cluster.Stream {
				fmt.Println("suspend supervisor")
				clusterWrapper.SuspendSupervisors()
			}
		}
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.Send(request, response)
}

func (provisionerThread *ProvisionerThread) ProcessProvisionRequest(request *threads.ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)

	if !found {
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision clusters from a mounted module
	// - if a module is unmounted, it is not meant to be operational
	if !moduleWrapper.IsMounted() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; it's module was not mounted\n", utils.Green, request.ModuleName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

	if !found {
		provisionerThread.logger.Warnf("%s[%s]%s Cluster does not exist\n", utils.Green, request.ClusterName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision mounted etl processes
	// - if a cluster is unmounted, even if the module is mounted, it is not meant to be operational
	if !clusterWrapper.IsMounted() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; cluster was not mounted\n", utils.Green, request.ClusterName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision batch etl processes
	// - stream processes are meant to be run by the system when mounted or unmounted
	if (request.Source == threads.Http) && clusterWrapper.IsStream() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; it's a stream process\n", utils.Green, request.ModuleName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	provisionerThread.logger.Printf("%s[%s]%s Provisioning cluster in module %s\n", utils.Green, request.ClusterName, utils.Reset, request.ModuleName)

	// if the operator does not specify a config to use, the system shall use the cluster identifier name
	// to find a default config that should be located in the database thread
	if request.Metadata.ConfigName == "" {
		request.Metadata.ConfigName = request.ClusterName
	}

	cnf, configFound := GetConfigFromDatabase(provisionerThread.C7, provisionerThread.databaseResponseTable, request.ModuleName, request.Metadata.ConfigName)
	if !configFound {
		// the config was either never created or deleted from the database.
		// INSTEAD of continuing, the node should inform the user that the client cannot use the config they want
		response := &threads.ProvisionerResponse{Success: false, Description: "config not found", Nonce: request.Nonce}
		provisionerThread.Send(request, response)
		provisionerThread.requestWg.Done()
		return
	}

	var supervisorInstance *supervisor.Supervisor
	if configFound {
		provisionerThread.logger.Printf("%s[%s]%s Initializing cluster supervisor from config\n", utils.Green, request.ClusterName, utils.Reset)
		cnf.Print()
		supervisorInstance = clusterWrapper.CreateSupervisor(request.Metadata.Other, cnf)
	} else {
		provisionerThread.logger.Printf("%s[%s]%s Initializing cluster supervisor\n", utils.Green, request.ClusterName, utils.Reset)
		supervisorInstance = clusterWrapper.CreateSupervisor(request.Metadata.Other)
	}

	provisionerThread.logger.Printf("%s[%s]%s Supervisor(%d) registered to cluster(%s)\n", utils.Green, request.ClusterName, utils.Reset, supervisorInstance.Id, request.ClusterName)

	response := &threads.ProvisionerResponse{
		Nonce:        request.Nonce,
		Success:      true,
		Cluster:      request.ClusterName,
		SupervisorId: supervisorInstance.Id,
	}
	provisionerThread.Send(request, response)

	provisionerThread.logger.Printf("%s[%s]%s Cluster Running\n", utils.Green, request.ClusterName, utils.Reset)

	go func() {

		// block until the supervisor completes
		response := supervisorInstance.Start()

		// don't send the statistics of the cluster to the database unless an Identifier has been
		// given to the cluster for grouping purposes
		if len(supervisorInstance.Config.Identifier) != 0 {
			// saves statistics to the database thread
			dbRequest := threads.DatabaseRequest{
				Action:  threads.DatabaseStore,
				Origin:  threads.Provisioner,
				Cluster: supervisorInstance.Config.Identifier,
				Data:    response,
			}
			provisionerThread.C7 <- dbRequest

			// sends a completion message to the messenger thread to write to a log file or send an email regarding completion
			msgRequest := threads.MessengerRequest{
				Action:  threads.MessengerClose,
				Cluster: supervisorInstance.Config.Identifier,
			}
			provisionerThread.C11 <- msgRequest

			// provide the console with output indicating that the cluster has completed
			// we already provide output when a cluster is provisioned, so it completes the state
			if GetConfigInstance().Debug {
				duration := time.Now().Sub(supervisorInstance.StartTime)
				provisionerThread.logger.Printf("%s[%s]%s Cluster transformations complete, took %dhr %dm %ds %dms %dus\n",
					utils.Green,
					supervisorInstance.Config.Identifier,
					utils.Reset,
					int(duration.Hours()),
					int(duration.Minutes()),
					int(duration.Seconds()),
					int(duration.Milliseconds()),
					int(duration.Microseconds()),
				)
			}
		}

		// let the provisioner thread decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-framework to shut down
		provisionerThread.requestWg.Done()
	}()
}

func (provisionerThread *ProvisionerThread) ProcessesIncomingDatabaseResponses(response threads.DatabaseResponse) {
	provisionerThread.databaseResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) ProcessIncomingCacheResponses(response threads.CacheResponse) {
	provisionerThread.cacheResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) Teardown() {
	provisionerThread.accepting = false

	provisionerThread.requestWg.Wait()
}
