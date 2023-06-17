package core

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/supervisor"
	"github.com/GabeCordo/etl/components/utils"
	"math/rand"
	"time"
)

func GetConfigFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName, clusterName string) (config cluster.Config, found bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseFetch,
		Type:    database.Config,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
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

func StoreConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName string, config cluster.Config) (success bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseStore,
		Type:    database.Config,
		Module:  moduleName,
		Cluster: config.Identifier,
		Data:    config,
		Nonce:   rand.Uint32(),
	}
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

func ReplaceConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName string, config cluster.Config) (success bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseReplace,
		Type:    database.Config,
		Module:  moduleName,
		Cluster: config.Identifier,
		Data:    config,
		Nonce:   rand.Uint32(),
	}
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

func ClusterMount(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{
		Action:      ProvisionerMount,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
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

	return !timeout && provisionerResponse.Success
}

func ClusterUnMount(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{
		Action:      ProvisionerUnMount,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
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

	return !timeout && provisionerResponse.Success
}

func DynamicallyDeleteCluster(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, clusterName string) (success bool) {

	provisionerThreadRequest := ProvisionerRequest{
		Action:      ProvisionerDynamicDelete,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
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

	return timeout || provisionerResponse.Success
}

func SupervisorProvision(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, moduleName, clusterName string, config ...string) (supervisorId uint64, success bool, description string) {

	provisionerThreadRequest := ProvisionerRequest{
		Action:      ProvisionerProvision,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
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
		return 0, false, provisionerResponse.Description
	} else {
		return provisionerResponse.SupervisorId, provisionerResponse.Success, provisionerResponse.Description
	}
}

func ClusterList(moduleName string) (clusters map[string]bool, success bool) {

	provisionerInstance := GetProvisionerInstance()

	clusters = make(map[string]bool, 0)

	moduleWrapper, found := provisionerInstance.GetModule(moduleName)
	if !found {
		return nil, false
	}

	mounts := moduleWrapper.GetClustersData()
	for identifier, isMounted := range mounts {
		clusters[identifier] = isMounted
	}

	return clusters, true
}

func SupervisorLookup(moduleName, clusterName string, supervisorId uint64) (supervisorInstance *supervisor.Supervisor, success bool) {

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(moduleName)
	if !found {
		fmt.Println("module not found")
		return nil, false
	}

	clusterWrapper, found := moduleWrapper.GetCluster(clusterName)
	if !found {
		fmt.Println("cluster not found")
		return nil, false
	}

	supervisorInstance, found = clusterWrapper.FindSupervisor(supervisorId)
	if !found {
		fmt.Println("supervisor not found")
		return nil, false
	}

	fmt.Println("returning supervisor")

	return supervisorInstance, found
}

func FindStatistics(pipe chan<- DatabaseRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (entries []database.Entry, found bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseFetch,
		Type:    database.Statistic,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
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

func PingNodeChannels(logger *utils.Logger, databasePipe chan<- DatabaseRequest, databaseResponseTable *utils.ResponseTable, provisionerPipe chan<- ProvisionerRequest, provisionerResponseTable *utils.ResponseTable) (success bool) {

	databasePingRequest := DatabaseRequest{
		Action: DatabaseUpperPing,
		Nonce:  rand.Uint32(),
	}
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
		logger.Println("received ping over C2")
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
		logger.Println("received ping over C6")
	}

	return true
}

func RegisterModule(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, modulePath string) (success bool, description string) {

	request := ProvisionerRequest{
		Action:     ProvisionerModuleLoad,
		ModulePath: modulePath,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	provisionerTimeout := false
	var provisionerResponse ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(request.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	success = !provisionerTimeout && provisionerResponse.Success
	return success, provisionerResponse.Description
}

func DeleteModule(pipe chan<- ProvisionerRequest, responseTable *utils.ResponseTable, moduleName string) (success bool, description string) {

	request := ProvisionerRequest{
		Action:     ProvisionerModuleDelete,
		ModuleName: moduleName,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	provisionerTimeout := false
	var provisionerResponse ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(request.Nonce); found {
			provisionerResponse = (responseEntry).(ProvisionerResponse)
			break
		}
	}

	success = !provisionerTimeout && provisionerResponse.Success
	return success, provisionerResponse.Description
}

func ToggleDebugMode(logger *utils.Logger) (description string) {

	config := GetConfigInstance()
	config.Debug = !config.Debug

	if config.Debug {
		description = "debug mode activated"
		logger.Println("remote change: debug mode ON")
	} else {
		description = "debug mode disabled"
		logger.Println("remote change: debug mode OFF")
	}

	return description
}
