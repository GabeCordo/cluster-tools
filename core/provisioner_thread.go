package core

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/module"
	"github.com/GabeCordo/etl/components/provisioner"
	"github.com/GabeCordo/etl/components/supervisor"
	"github.com/GabeCordo/etl/components/utils"
	"log"
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
	GetProvisionerInstance() // create the supervisor if it doesn't exist
}

func (provisionerThread *ProvisionerThread) Start() {
	go func() {
		// request coming from http_server
		for request := range provisionerThread.C5 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessIncomingRequests(&request)
		}

		provisionerThread.wg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C8 {
			if !provisionerThread.accepting {
				break
			}

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessesIncomingDatabaseResponses(response)
		}

		provisionerThread.wg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C10 {
			if !provisionerThread.accepting {
				break
			}

			// if this doesn't spawn its own thread we can be left waiting
			provisionerThread.ProcessIncomingCacheResponses(response)
		}

		provisionerThread.wg.Wait()
	}()

	provisionerThread.wg.Wait()
}

func (provisionerThread *ProvisionerThread) ProcessIncomingRequests(request *ProvisionerRequest) {

	if request.Action == ProvisionerMount {
		provisionerThread.ProcessMountRequest(request)
	} else if request.Action == ProvisionerUnMount {
		provisionerThread.ProcessUnMountRequest(request)
	} else if request.Action == ProvisionerProvision {
		provisionerThread.ProcessProvisionRequest(request)
	} else if request.Action == ProvisionerTeardown {
		// TODO - not implemented
	} else if request.Action == ProvisionerLowerPing {
		provisionerThread.ProcessPingProvisionerRequest(request)
	} else if request.Action == ProvisionerModuleLoad {
		provisionerThread.RegisterModule(request)
	} else if request.Action == ProvisionerModuleDelete {
		provisionerThread.ProcessDeleteModule(request)
	}
}

func (provisionerThread *ProvisionerThread) RegisterModule(request *ProvisionerRequest) {

	log.Printf("registering module at %s\n", request.ModulePath)

	remoteModule, err := module.NewRemoteModule(request.ModulePath)
	if err != nil {
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "cannot find remote module"}
		provisionerThread.wg.Done()
		return
	}

	moduleInstance, err := remoteModule.Get()
	if err != nil {
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "module built with older version"}
		provisionerThread.wg.Done()
		return
	}

	success := GetProvisionerInstance().AddModule(moduleInstance)
	if !success {
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "a module with that identifier already exists or is corrupt"}
		provisionerThread.wg.Done()
		return
	}

	// if one of the clusters in the module is marked as mounted, then the module should be mounted itself
	for _, export := range moduleInstance.Config.Exports {
		if export.StaticMount {
			moduleWrapper, _ := GetProvisionerInstance().GetModule(moduleInstance.Config.Identifier)
			moduleWrapper.Mount()
			break
		}
	}

	// note: the module wrapper should already be defined so there is no need to validate
	moduleWrapper, _ := GetProvisionerInstance().GetModule(moduleInstance.Config.Identifier)

	// REGISTER ANY HELPERS TO CLUSTERS THAT HAVE DEFINED THEM WITHIN THE STRUCT
	for _, clusterWrapper := range moduleWrapper.GetClusters() {

		clusterImplementation := clusterWrapper.GetClusterImplementation()

		reflectedClusterImpl := reflect.ValueOf(clusterImplementation)

		helperField := reflect.Indirect(reflectedClusterImpl).FieldByName("Helper")
		if helperField.IsValid() && helperField.CanSet() {
			if helper, err := NewHelper(provisionerThread.C9, provisionerThread.C11); err == nil {
				helperField.Set(reflect.ValueOf(helper))
			}
		}
	}

	// REGISTER EACH CONFIG FROM THE MODULE FILE TO THE DATABASE THREAD
	for _, export := range moduleInstance.Config.Exports {

		config := cluster.Config{
			Identifier:                  export.Cluster,
			Mode:                        export.Config.OnCrash,
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

		request := DatabaseRequest{
			Action:  DatabaseStore,
			Nonce:   rand.Uint32(),
			Origin:  Provisioner,
			Type:    database.Config,
			Cluster: export.Cluster,
			Data:    config,
		}
		provisionerThread.C7 <- request

		timeout := false
		var response DatabaseResponse

		timestamp := time.Now()
		for {
			if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
				timeout = true
				break
			}

			if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(request.Nonce); found {
				response = (responseEntry).(DatabaseResponse)
				break
			}
		}

		if timeout || !response.Success {
			provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "could not save config"}
			provisionerThread.wg.Done()
			return
		}
	}

	log.Printf("[provisioner] dynamically loaded module %s\n", moduleInstance.Config.Identifier)

	provisionerThread.C6 <- ProvisionerResponse{Success: true, Nonce: request.Nonce, Description: "module registered"}
	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessDeleteModule(request *ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	var response ProvisionerResponse = ProvisionerResponse{Nonce: request.Nonce}
	if deleted, _, found := provisionerInstance.DeleteModule(request.ModuleName); found {
		response.Success = true
		if deleted {
			response.Description = "module deleted"
		} else {
			response.Description = "module marked for deletion, a cluster is likely running right now, try later"
		}
	} else {
		response.Success = false
		response.Description = "module not found"
	}

	provisionerThread.C6 <- response
	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessPingProvisionerRequest(request *ProvisionerRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_provisioner] received ping over C5")
	}

	databaseRequest := DatabaseRequest{Action: DatabaseLowerPing, Nonce: rand.Uint32()}
	provisionerThread.C7 <- databaseRequest

	databasePingTimeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			databasePingTimeout = true
			break
		}

		if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	if databasePingTimeout || !databaseResponse.Success {
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	if GetConfigInstance().Debug {
		log.Println("[etl_provisioner] received ping over C8")
	}

	cacheRequest := CacheRequest{Action: CacheLowerPing, Nonce: rand.Uint32()}
	provisionerThread.C9 <- cacheRequest

	cachePingTimeout := false
	var cacheResponse CacheResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			cachePingTimeout = true
			break
		}

		if response, found := provisionerThread.cacheResponseTable.Lookup(cacheRequest.Nonce); found {
			cacheResponse = (response).(CacheResponse)
			break
		}
	}

	if cachePingTimeout || !cacheResponse.Success {
		log.Println("[etl_provisioner] failed to receive ping over C10")
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	if GetConfigInstance().Debug {
		log.Println("[etl_provisioner] received ping over C10")
	}

	provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: true}

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessMountRequest(request *ProvisionerRequest) {

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.wg.Done()
		return
	}

	if !moduleWrapper.IsMounted() {
		moduleWrapper.Mount()

		if GetConfigInstance().Debug {
			log.Printf("%s[%s]%s Mounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.wg.Done()
			return
		}

		if !clusterWrapper.IsMounted() {
			clusterWrapper.Mount()

			if GetConfigInstance().Debug {
				log.Printf("%s[%s]%s Mounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}
		}
	}

	provisionerThread.C6 <- ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessUnMountRequest(request *ProvisionerRequest) {

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.wg.Done()
		return
	}

	if moduleWrapper.IsMounted() {
		moduleWrapper.UnMount()

		if GetConfigInstance().Debug {
			log.Printf("%s[%s]%s UnMounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			provisionerThread.C6 <- ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.wg.Done()
			return
		}

		if clusterWrapper.IsMounted() {
			clusterWrapper.UnMount()

			if GetConfigInstance().Debug {
				log.Printf("%s[%s]%s UnMounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}
		}
	}

	provisionerThread.C6 <- ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessProvisionRequest(request *ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)

	if !found {
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	if !moduleWrapper.IsMounted() {
		log.Printf("%s[%s]%s Could not provision cluster; it's module was not mounted\n", utils.Green, request.ModuleName, utils.Reset)
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

	if !found {
		log.Printf("%s[%s]%s Cluster does not exist\n", utils.Green, request.ClusterName, utils.Reset)
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	if !clusterWrapper.IsMounted() {
		log.Printf("%s[%s]%s Could not provision cluster; cluster was not mounted\n", utils.Green, request.ClusterName, utils.Reset)
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	log.Printf("%s[%s]%s Provisioning cluster in module %s\n", utils.Green, request.ClusterName, utils.Reset, request.ModuleName)

	// if the operator does not specify a config to use, the system shall use the cluster identifier name
	// to find a default config that should be located in the database thread
	if request.Config == "" {
		request.Config = request.ClusterName
	}

	config, configFound := GetConfigFromDatabase(provisionerThread.C7, provisionerThread.databaseResponseTable, request.Config)
	config.Print()
	fmt.Println(configFound)
	if !configFound {
		// the config was either never created or deleted from the database.
		// INSTEAD of continuing, the node should inform the user that the client cannot use the config they want
		provisionerThread.C6 <- ProvisionerResponse{Success: false, Description: "config not found", Nonce: request.Nonce}
		provisionerThread.wg.Done()
		return
	}

	var supervisorInstance *supervisor.Supervisor
	if configFound {
		log.Printf("%s[%s]%s Initializing cluster supervisor from config\n", utils.Green, request.ClusterName, utils.Reset)
		config.Print()
		supervisorInstance = clusterWrapper.CreateSupervisor(config)
	} else {
		log.Printf("%s[%s]%s Initializing cluster supervisor\n", utils.Green, request.ClusterName, utils.Reset)
		supervisorInstance = clusterWrapper.CreateSupervisor()
	}

	log.Printf("%s[%s]%s Supervisor(%d) registered to cluster(%s)\n", utils.Green, request.ClusterName, utils.Reset, supervisorInstance.Id, request.ClusterName)

	provisionerThread.C6 <- ProvisionerResponse{
		Nonce:        request.Nonce,
		Success:      true,
		Cluster:      request.ClusterName,
		SupervisorId: supervisorInstance.Id,
	}

	log.Printf("%s[%s]%s Cluster Running\n", utils.Green, request.ClusterName, utils.Reset)

	go func() {

		// block until the supervisor completes
		supervisorInstance.Print()
		response := supervisorInstance.Start()

		// don't send the statistics of the cluster to the database unless an Identifier has been
		// given to the cluster for grouping purposes
		if len(supervisorInstance.Config.Identifier) != 0 {
			// saves statistics to the database thread
			dbRequest := DatabaseRequest{Action: DatabaseStore, Origin: Provisioner, Cluster: supervisorInstance.Config.Identifier, Data: response}
			provisionerThread.C7 <- dbRequest

			// sends a completion message to the messenger thread to write to a log file or send an email regarding completion
			msgRequest := MessengerRequest{Action: MessengerClose, Cluster: supervisorInstance.Config.Identifier}
			provisionerThread.C11 <- msgRequest

			// provide the console with output indicating that the cluster has completed
			// we already provide output when a cluster is provisioned, so it completes the state
			if GetConfigInstance().Debug {
				duration := time.Now().Sub(supervisorInstance.StartTime)
				log.Printf("%s[%s]%s Cluster transformations complete, took %dhr %dm %ds %dms %dus\n",
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
		provisionerThread.wg.Done()
	}()
}

func (provisionerThread *ProvisionerThread) ProcessesIncomingDatabaseResponses(response DatabaseResponse) {
	provisionerThread.databaseResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) ProcessIncomingCacheResponses(response CacheResponse) {
	provisionerThread.cacheResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) Teardown() {
	provisionerThread.accepting = false

	provisionerThread.wg.Wait()
}
