package core

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"log"
	"math/rand"
	"time"
)

/**
 * Cluster Request Handlers
 */

func (httpThread *HttpThread) ClusterMount(cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerMount}
	httpThread.C5 <- provisionerThreadRequest

	return true
}

func (httpThread *HttpThread) ClusterUnMount(cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerUnMount}
	httpThread.C5 <- provisionerThreadRequest

	return true
}

func (httpThread *HttpThread) ClusterProvision(cluster string) (supervisorId uint64, success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: ProvisionerProvision}
	httpThread.C5 <- provisionerThreadRequest

	timeout := false
	var provisionerResponse ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := httpThread.provisionerResponseTable.Lookup(provisionerThreadRequest.Nonce); found {
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

func (httpThread *HttpThread) ClusterList() (clusters map[string]bool, success bool) {

	provisionerInstance := GetProvisionerInstance()

	clusters = make(map[string]bool, 0)

	mounts := provisionerInstance.Mounts()
	for identifier, isMounted := range mounts {
		clusters[identifier] = isMounted
	}

	return clusters, true
}

/**
 * Supervisor Request Handlers
 */

func (httpThread *HttpThread) SupervisorLookup(clusterId string, supervisorId uint64) (supervisor cluster.Supervisor, success bool) {

	provisionerInstance := GetProvisionerInstance()

	clusterRegistry, found := provisionerInstance.GetRegistry(clusterId)
	if !found {
		return cluster.Supervisor{}, false
	}

	supervisorInstance, found := clusterRegistry.GetSupervisor(supervisorId)
	if !found {
		return cluster.Supervisor{}, false
	}

	return *supervisorInstance, true
}

/**
 * Config Request Handlers
 */

func (httpThread *HttpThread) GetConfig(clusterName string) (config cluster.Config, found bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseFetch, Type: database.Config, Nonce: rand.Uint32(), Cluster: clusterName}
	httpThread.C1 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := httpThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
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

func (httpThread *HttpThread) StoreConfig(config cluster.Config) (success bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseStore, Type: database.Config, Nonce: rand.Uint32(), Cluster: config.Identifier, Data: config}
	httpThread.C1 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := httpThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(DatabaseResponse)
			break
		}
	}

	return timeout || databaseResponse.Success
}

/**
 * Statistics Request Handlers
 */

func (httpThread *HttpThread) FindStatistics(clusterName string) (entries []database.Entry, found bool) {

	databaseRequest := DatabaseRequest{Action: DatabaseFetch, Type: database.Statistic, Nonce: rand.Uint32(), Cluster: clusterName}
	httpThread.C1 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := httpThread.databaseResponseTable.Lookup(databaseRequest.Nonce); found {
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

/**
 * Debug Request Handlers
 */

func (httpThread *HttpThread) ShutdownNode(body DebugJSONBody) (response []byte, success bool) {
	httpThread.Interrupt <- Shutdown

	return nil, true
}

func (httpThread *HttpThread) PingNodeChannels() (success bool) {

	databasePingRequest := DatabaseRequest{Action: DatabaseUpperPing, Nonce: rand.Uint32()}
	httpThread.C1 <- databasePingRequest

	databaseTimeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			databaseTimeout = true
			break
		}

		if responseEntry, found := httpThread.databaseResponseTable.Lookup(databasePingRequest.Nonce); found {
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
	httpThread.C5 <- provisionerPingRequest

	provisionerTimeout := false
	var provisionerResponse ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > 2.0 {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := httpThread.provisionerResponseTable.Lookup(provisionerPingRequest.Nonce); found {
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
