package core

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/supervisor"
	"github.com/GabeCordo/etl/components/utils"
	"math/rand"
	"time"
)

func GetConfigFromDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName, clusterName string) (conf cluster.Config, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	if timeout || !databaseResponse.Success {
		return cluster.Config{}, false
	} else {
		return (databaseResponse.Data).(cluster.Config), true
	}
}

func StoreConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName string, config cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseStore,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: config.Identifier,
		Data:    config,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	return timeout || databaseResponse.Success
}

func ReplaceConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable, moduleName string, config cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseReplace,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: config.Identifier,
		Data:    config,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	return timeout || databaseResponse.Success
}

func ClusterMount(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerMount,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
			break
		}
	}

	return !timeout && provisionerResponse.Success
}

func ClusterUnMount(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerUnMount,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
			break
		}
	}

	return !timeout && provisionerResponse.Success
}

func DynamicallyDeleteCluster(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable, clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerDynamicDelete,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
			break
		}
	}

	return timeout || provisionerResponse.Success
}

func SupervisorProvision(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string, meta map[string]string, config ...string) (supervisorId uint64, success bool, description string) {

	// there is a possibility the user never passed an args value to the HTTP endpoint,
	// so we need to replace it with and empty arry
	if meta == nil {
		meta = make(map[string]string)
	}
	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerProvision,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
		MetaData:    meta,
	}
	if len(config) > 0 {
		provisionerThreadRequest.Config = config[0]
	}
	pipe <- provisionerThreadRequest

	timeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(provisionerThreadRequest.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
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
		return nil, false
	}

	clusterWrapper, found := moduleWrapper.GetCluster(clusterName)
	if !found {
		return nil, false
	}

	supervisorInstance, found = clusterWrapper.FindSupervisor(supervisorId)
	if !found {
		return nil, false
	}

	return supervisorInstance, found
}

func FindStatistics(pipe chan<- threads.DatabaseRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (entries []database.Entry, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.SupervisorStatistic,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	timeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			timeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(databaseRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	if timeout || !databaseResponse.Success {
		return nil, false
	} else {
		return (databaseResponse.Data).([]database.Entry), true
	}
}

func ShutdownNode(pipe chan<- threads.InterruptEvent) (response []byte, success bool) {
	pipe <- threads.Shutdown
	return nil, true
}

func PingNodeChannels(logger *utils.Logger, databasePipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable, provisionerPipe chan<- threads.ProvisionerRequest, provisionerResponseTable *utils.ResponseTable) (success bool) {

	databasePingRequest := threads.DatabaseRequest{
		Action: threads.DatabaseUpperPing,
		Nonce:  rand.Uint32(),
	}
	databasePipe <- databasePingRequest

	databaseTimeout := false
	var databaseResponse threads.DatabaseResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			databaseTimeout = true
			break
		}

		if responseEntry, found := databaseResponseTable.Lookup(databasePingRequest.Nonce); found {
			databaseResponse = (responseEntry).(threads.DatabaseResponse)
			break
		}
	}

	if databaseTimeout || !databaseResponse.Success {
		return false
	}

	if GetConfigInstance().Debug {
		logger.Println("received ping over C2")
	}

	provisionerPingRequest := threads.ProvisionerRequest{
		Action: threads.ProvisionerLowerPing,
		Nonce:  rand.Uint32(),
	}
	provisionerPipe <- provisionerPingRequest

	provisionerTimeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := provisionerResponseTable.Lookup(provisionerPingRequest.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
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

func RegisterModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable, modulePath string) (success bool, description string) {

	request := threads.ProvisionerRequest{
		Action:     threads.ProvisionerModuleLoad,
		ModulePath: modulePath,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	provisionerTimeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(request.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
			break
		}
	}

	success = !provisionerTimeout && provisionerResponse.Success
	return success, provisionerResponse.Description
}

func DeleteModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable, moduleName string) (success bool, description string) {

	request := threads.ProvisionerRequest{
		Action:     threads.ProvisionerModuleDelete,
		ModuleName: moduleName,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	provisionerTimeout := false
	var provisionerResponse threads.ProvisionerResponse

	timestamp2 := time.Now()
	for {
		if time.Now().Sub(timestamp2).Seconds() > GetConfigInstance().MaxWaitForResponse {
			provisionerTimeout = true
			break
		}

		if responseEntry, found := responseTable.Lookup(request.Nonce); found {
			provisionerResponse = (responseEntry).(threads.ProvisionerResponse)
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
