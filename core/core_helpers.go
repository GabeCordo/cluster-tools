package core

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/supervisor"
	"github.com/GabeCordo/etl/components/utils"
	"log"
	"math/rand"
	"time"
)

func GetConfigFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, clusterName string) (config cluster.Config, found bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseFetch, Type: database.Config, Nonce: rand.Uint32(), Cluster: clusterName}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	if timeout || !databaseResponse.Success {
		return cluster.Config{}, false
	} else {
		return *(databaseResponse.Data).(*cluster.Config), true
	}
}

func StoreConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, config cluster.Config) (success bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseStore, Type: database.Config, Nonce: rand.Uint32(), Cluster: config.Identifier, Data: config}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	return timeout || databaseResponse.Success
}

func ReplaceConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, config cluster.Config) (success bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseReplace, Type: database.Config, Nonce: rand.Uint32(), Cluster: config.Identifier, Data: config}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	return timeout || databaseResponse.Success
}

func ClusterMount(pipe chan<- ProvisionerRequest, cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerMount}
	pipe <- provisionerThreadRequest

	return true
}

func ClusterUnMount(pipe chan<- ProvisionerRequest, cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerUnMount}
	pipe <- provisionerThreadRequest

	return true
}

func DynamicallyRegisterCluster(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, clusterName, sharedObjectPath string) (success bool, description string) {

	provisionerThreadRequest := ProvisionerRequest{Action: ProvisionerDynamicLoad, Nonce: rand.Uint32(), Cluster: clusterName, Path: sharedObjectPath}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	return timeout || provisionerResponse.Success, provisionerResponse.Description
}

func DynamicallyDeleteCluster(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, clusterName string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Action: ProvisionerDynamicDelete, Nonce: rand.Uint32(), Cluster: clusterName}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	return timeout || provisionerResponse.Success
}

func SupervisorProvision(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, cluster string, config ...string) (supervisorId uint64, success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerProvision}
	if len(config) > 0 {
		provisionerThreadRequest.Config = config[0]
	}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	if timeout {
		return 0, false
	} else {
		return provisionerResponse.SupervisorId, true
	}
}

func ClusterList() (clusters map[string]bool, success bool) {

	provisionerInstance := GetProvisionerInstance()

	clusters = make(map[string]bool, 0)

	mounts := provisionerInstance.Mounts()
	for identifier, isMounted := range mounts {
		clusters[identifier] = isMounted
	}

	return clusters, true
}

func SupervisorLookup(clusterId string, supervisorId uint64) (supervisorInstance *supervisor.Supervisor, success bool) {

	provisionerInstance := GetProvisionerInstance()

	clusterRegistry, found := provisionerInstance.GetRegistry(clusterId)
	if !found {
		return nil, false
	}

	return clusterRegistry.GetSupervisor(supervisorId)
}

func FindStatistics(pipe chan<- DatabaseRequest, responseTable *utils.ResponseTable, clusterName string) (entries []database.Entry, found bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseFetch, Type: database.Statistic, Nonce: rand.Uint32(), Cluster: clusterName}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	if timeout || !databaseResponse.Success {
		return nil, false
	} else {
		return (databaseResponse.Data).([]database.Entry), true
	}
}

func ShutdownNode(pipe chan<- InterruptEvent) (response []byte, success bool) {
	pipe <- Shutdown
	return nil, true
}

func PingNodeChannels(databasePipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, provisionerPipe chan<- ProvisionerRequest, provisionerResponseTable *utils.ResponseTable) (success bool) {

	databasePingRequest := DatabaseRequest{Action: DatabaseUpperPing, Nonce: rand.Uint32()}
	databasePipe <- databasePingRequest

	databaseTimeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			databaseTimeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databasePingRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	if databaseTimeout || !databaseResponse.Success {
		return false
	}

	if GetConfigInstance().Debug {
		log.Println("[etl_http] received ping over C2")
	}

	provisionerPingRequest := ProvisionerRequest{Action: ProvisionerLowerPing, Nonce: rand.Uint32()}
	provisionerPipe <- provisionerPingRequest

	provisionerTimeout := false
	var provisionerResponse ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := provisionerResponseTable.Lookup(provisionerPingRequest.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	if provisionerTimeout || !provisionerResponse.Success {
		return false
	}

	if GetConfigInstance().Debug {
		log.Println("[etl_provisioner] received ping over C6")
	}

	return true
}
