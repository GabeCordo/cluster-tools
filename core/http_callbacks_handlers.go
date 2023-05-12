package core

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"math/rand"
	"time"
)

/**
 * Cluster Request Handlers
 */

func (httpThread *HttpThread) ClusterMount(cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: Mount}
	httpThread.C5 <- provisionerThreadRequest

	return true
}

func (httpThread *HttpThread) ClusterUnMount(cluster string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: UnMount}
	httpThread.C5 <- provisionerThreadRequest

	return true
}

func (httpThread *HttpThread) ClusterProvision(cluster string) (supervisorId uint64, success bool) {

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: cluster, Action: Provision}
	httpThread.C5 <- provisionerThreadRequest

	timeout := false
	var provisionerResponse ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := GetProvisionerResponseTable().Lookup(provisionerThreadRequest.Nonce); found {
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

	databaseRequest := DatabaseRequest{Action: Fetch, Type: database.Config, Nonce: rand.Uint32(), Cluster: clusterName}
	httpThread.C1 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := GetDatabaseResponseTable().Lookup(databaseRequest.Nonce); found {
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

	fmt.Println(config)

	databaseRequest := DatabaseRequest{Action: Store, Type: database.Config, Nonce: rand.Uint32(), Cluster: config.Identifier, Data: config}
	httpThread.C1 <- databaseRequest

	fmt.Println(databaseRequest)

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := GetDatabaseResponseTable().Lookup(databaseRequest.Nonce); found {
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

	databaseRequest := DatabaseRequest{Action: Fetch, Type: database.Statistic, Nonce: rand.Uint32(), Cluster: clusterName}
	httpThread.C1 <- databaseRequest

	timeout := false
	var databaseResponse DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			timeout = true
			break
		}

		if responseEntry, found := GetDatabaseResponseTable().Lookup(databaseRequest.Nonce); found {
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
