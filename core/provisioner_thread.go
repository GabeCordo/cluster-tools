package core

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/provisioner"
	"github.com/GabeCordo/etl/components/supervisor"
	"github.com/GabeCordo/etl/components/utils"
	"log"
	"math/rand"
	"plugin"
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
	provisionerInstance := GetProvisionerInstance() // create the supervisor if it doesn't exist

	GetProvisionerMemoryInstance() // create the instance

	// auto-mounting is supported within the etl Config; if a cluster Identifier is added
	// to the config under 'auto-mount', it is added to the map of Operational functions
	for _, identifier := range GetConfigInstance().AutoMount {
		provisionerInstance.Mount(identifier)
	}
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
	} else if request.Action == ProvisionerDynamicLoad {
		provisionerThread.ProcessDynamicClusterLoad(request)
	} else if request.Action == ProvisionerDynamicDelete {
		provisionerThread.ProcessDynamicClusterDelete(request)
	}
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

	GetProvisionerInstance().Mount(request.Cluster)

	success := GetProvisionerInstance().IsMounted(request.Cluster)
	provisionerThread.C6 <- ProvisionerResponse{Success: success, Nonce: request.Nonce}

	if GetConfigInstance().Debug && success {
		log.Printf("%s[%s]%s Mounted cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessUnMountRequest(request *ProvisionerRequest) {

	GetProvisionerInstance().UnMount(request.Cluster)

	success := !GetProvisionerInstance().IsMounted(request.Cluster)
	provisionerThread.C6 <- ProvisionerResponse{Success: success, Nonce: request.Nonce}

	if GetConfigInstance().Debug && success {
		log.Printf("%s[%s]%s UnMounted cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessProvisionRequest(request *ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	if !provisionerInstance.IsMounted(request.Cluster) {
		log.Printf("%s[%s]%s Could not provision cluster; cluster was not mounted\n", utils.Green, request.Cluster, utils.Reset)
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	} else {
		log.Printf("%s[%s]%s Provisioning cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	_, ok := provisionerInstance.Function(request.Cluster)
	if !ok {
		log.Printf("%s[%s]%s There is a corrupted cluster in the supervisor\n", utils.Green, request.Cluster, utils.Reset)
		provisionerThread.C6 <- ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.wg.Done()
		return
	}

	// if the operator does not specify a config to use, the system shall use the cluster identifier name
	// to find a default config that should be located in the database thread
	if request.Config == "" {
		request.Config = request.Cluster
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

	registryInstance, _ := provisionerInstance.GetRegistry(request.Cluster)

	var supervisorInstance *supervisor.Supervisor
	if configFound {
		log.Printf("%s[%s]%s Initializing cluster supervisor from config\n", utils.Green, request.Cluster, utils.Reset)
		config.Print()
		supervisorInstance = registryInstance.CreateSupervisor(config)
	} else {
		log.Printf("%s[%s]%s Initializing cluster supervisor\n", utils.Green, request.Cluster, utils.Reset)
		supervisorInstance = registryInstance.CreateSupervisor()
	}

	log.Printf("%s[%s]%s Supervisor(%d) registered to cluster(%s)\n", utils.Green, request.Cluster, utils.Reset, supervisorInstance.Id, request.Cluster)

	provisionerThread.C6 <- ProvisionerResponse{
		Nonce:        request.Nonce,
		Success:      true,
		Cluster:      request.Cluster,
		SupervisorId: supervisorInstance.Id,
	}

	log.Printf("%s[%s]%s Cluster Running\n", utils.Green, request.Cluster, utils.Reset)

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

func (provisionerThread *ProvisionerThread) ProcessDynamicClusterLoad(request *ProvisionerRequest) {

	response := ProvisionerResponse{Nonce: request.Nonce}
	if (len(request.Path) == 0) || (len(request.Cluster) == 0) {
		response.Success = false
		response.Description = "missing dynamic path or cluster identifier"
		provisionerThread.wg.Done()
		return
	}

	if _, found := GetProvisionerInstance().Function(request.Cluster); found {
		response.Success = false
		response.Description = "a cluster with that identifier already exists"
		provisionerThread.C6 <- response
		provisionerThread.wg.Done()
		return
	}

	dynamicClusterPlugin, err := plugin.Open(request.Path)
	if err != nil {
		response.Description = err.Error()
		response.Success = false
		provisionerThread.C6 <- response
		provisionerThread.wg.Done()
		return
	}

	symbol, err := dynamicClusterPlugin.Lookup(request.Cluster)
	if err != nil {
		response.Description = err.Error()
		response.Success = false
		provisionerThread.C6 <- response
		provisionerThread.wg.Done()
		return
	}

	dynamicallyLoadedCluster := symbol.(cluster.Cluster)

	GetProvisionerInstance().Register(request.Cluster, dynamicallyLoadedCluster)

	log.Printf("[provisioner] dynamically registered cluster %s\n", request.Cluster)

	if request.Mount {
		GetProvisionerInstance().Mount(request.Cluster)
		log.Printf("[provisioner] dynamically mounted cluster %s\n", request.Cluster)
	}

	databaseRequest := DatabaseRequest{Action: DatabaseStore, Type: database.Config, Cluster: request.Cluster, Data: cluster.Config{
		Identifier:                  request.Cluster,
		Mode:                        cluster.DoNothing,
		StartWithNTransformClusters: 1,
		StartWithNLoadClusters:      1,
		ETChannelThreshold:          1,
		ETChannelGrowthFactor:       2,
		TLChannelThreshold:          1,
		TLChannelGrowthFactor:       2,
	}}

	provisionerThread.C7 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := provisionerThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	if timeout || !databaseResponse.Success {
		response.Success = false
		provisionerThread.C6 <- response
		provisionerThread.wg.Done()
		return
	}

	response.Success = true
	provisionerThread.C6 <- response

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessDynamicClusterDelete(request *ProvisionerRequest) {

	success := GetProvisionerInstance().UnRegister(request.Cluster)

	if GetConfigInstance().Debug && success {
		log.Printf("[provisioner] un-registered cluster %s\n", request.Cluster)
	} else if GetConfigInstance().Debug {
		log.Printf("[provisioner] failed to un-registered cluster %s\n", request.Cluster)
	}

	provisionerThread.C6 <- ProvisionerResponse{Success: success, Nonce: request.Nonce}

	provisionerThread.wg.Done()
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
