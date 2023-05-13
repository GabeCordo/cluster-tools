package core

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/utils"
	"log"
	"math/rand"
	"time"
)

const (
	DefaultHardTerminateTime = 30 // minutes
)

var provisioner *cluster.Provisioner

func GetProvisionerInstance() *cluster.Provisioner {

	if provisioner == nil {
		provisioner = cluster.NewProvisioner()
	}
	return provisioner
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

	if GetConfigInstance().Debug {
		log.Printf("%s[%s]%s Mounted cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessUnMountRequest(request *ProvisionerRequest) {

	GetProvisionerInstance().UnMount(request.Cluster)

	if GetConfigInstance().Debug {
		log.Printf("%s[%s]%s UnMounted cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	provisionerThread.wg.Done()
}

func (provisionerThread *ProvisionerThread) ProcessProvisionRequest(request *ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	if !provisionerInstance.IsMounted(request.Cluster) {
		log.Printf("%s[%s]%s Could not provision cluster; cluster was not mounted\n", utils.Green, request.Cluster, utils.Reset)
		provisionerThread.wg.Done()
		return
	} else {
		log.Printf("%s[%s]%s Provisioning cluster\n", utils.Green, request.Cluster, utils.Reset)
	}

	clstr, cnfg, register, ok := provisionerInstance.Function(request.Cluster)
	if !ok {
		log.Printf("%s[%s]%s There is a corrupted cluster in the supervisor\n", utils.Green, request.Cluster, utils.Reset)
		provisionerThread.wg.Done()
		return
	}

	var supervisor *cluster.Supervisor
	if cnfg == nil {
		log.Printf("%s[%s]%s Initializing cluster supervisor\n", utils.Green, request.Cluster, utils.Reset)
		supervisor = cluster.NewSupervisor(*clstr)
	} else {
		log.Printf("%s[%s]%s Initializing cluster supervisor from config\n", utils.Green, request.Cluster, utils.Reset)
		supervisor = cluster.NewCustomSupervisor(*clstr, cnfg)
	}
	log.Printf("%s[%s]%s Registering supervisor(%d) to cluster(%s)\n", utils.Green, request.Cluster, utils.Reset, supervisor.Id, request.Cluster)
	register.Register(supervisor)
	log.Printf("%s[%s]%s Supervisor(%d) registered to cluster(%s)\n", utils.Green, request.Cluster, utils.Reset, supervisor.Id, request.Cluster)

	provisionerThread.C6 <- ProvisionerResponse{
		Nonce:        request.Nonce,
		Success:      true,
		Cluster:      request.Cluster,
		SupervisorId: supervisor.Id,
	}

	log.Printf("%s[%s]%s Cluster Running\n", utils.Green, request.Cluster, utils.Reset)

	go func() {
		var response *cluster.Response

		c := make(chan struct{})
		go func() {
			defer close(c)
			response = supervisor.Start()
		}()

		// block until the supervisor completes
		<-c

		// don't send the statistics of the cluster to the database unless an Identifier has been
		// given to the cluster for grouping purposes
		if len(supervisor.Config.Identifier) != 0 {
			// saves statistics to the database thread
			dbRequest := DatabaseRequest{Action: DatabaseStore, Origin: Provisioner, Cluster: supervisor.Config.Identifier, Data: response}
			provisionerThread.C7 <- dbRequest

			// sends a completion message to the messenger thread to write to a log file or send an email regarding completion
			msgRequest := MessengerRequest{Action: MessengerClose, Cluster: supervisor.Config.Identifier}
			provisionerThread.C11 <- msgRequest

			// provide the console with output indicating that the cluster has completed
			// we already provide output when a cluster is provisioned, so it completes the state
			if GetConfigInstance().Debug {
				duration := time.Now().Sub(supervisor.StartTime)
				log.Printf("%s[%s]%s Cluster transformations complete, took %dhr %dm %ds %dms %dus\n",
					utils.Green,
					supervisor.Config.Identifier,
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
