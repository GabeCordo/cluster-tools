package common

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/components/database"
	"github.com/GabeCordo/etl/core/components/provisioner"
	"github.com/GabeCordo/etl/core/components/supervisor"
	"github.com/GabeCordo/etl/core/utils"
	"math/rand"
)

func GetConfigFromDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName, clusterName string) (conf cluster.Config, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return cluster.Config{}, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return cluster.Config{}, false
	}

	configs := (databaseResponse.Data).([]cluster.Config)
	if len(configs) != 1 {
		return cluster.Config{}, false
	}

	return configs[0], true
}

func GetConfigsFromDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string) (configs []cluster.Config, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action: threads.DatabaseFetch,
		Type:   threads.ClusterConfig,
		Module: moduleName,
		Nonce:  rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	configs = (databaseResponse.Data).([]cluster.Config)
	return configs, true
}

func StoreConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string, cfg cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseStore,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func ReplaceConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string, cfg cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseReplace,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func DeleteConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName, configName string) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseDelete,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: configName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func ClusterMount(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerMount,
		Source:      threads.Http,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	data, didTimeout := utils.SendAndWait(responseTable, provisionerThreadRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)
	return provisionerResponse.Success
}

func ClusterUnMount(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerUnMount,
		Source:      threads.Http,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	data, didTimeout := utils.SendAndWait(responseTable, provisionerThreadRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)
	return provisionerResponse.Success
}

func DynamicallyDeleteCluster(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	clusterName string) (success bool) {

	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerDynamicDelete,
		Source:      threads.Http,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	data, didTimeout := utils.SendAndWait(responseTable, provisionerThreadRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)
	return provisionerResponse.Success
}

func SupervisorProvision(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string, meta map[string]string, cfg ...string) (supervisorId uint64, success bool, description string) {

	// there is a possibility the user never passed an args value to the HTTP endpoint,
	// so we need to replace it with and empty arry
	if meta == nil {
		meta = make(map[string]string)
	}
	provisionerThreadRequest := threads.ProvisionerRequest{
		Action:      threads.ProvisionerProvision,
		Source:      threads.Http,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Metadata: threads.ProvisionerMetadata{
			Other: meta,
		},
		Nonce: rand.Uint32(),
	}
	if len(cfg) > 0 {
		provisionerThreadRequest.Metadata.ConfigName = cfg[0]
	}
	pipe <- provisionerThreadRequest

	data, didTimeout := utils.SendAndWait(responseTable, provisionerThreadRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return 0, false, "timeout"
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)
	return provisionerResponse.SupervisorId, provisionerResponse.Success, provisionerResponse.Description
}

func GetClusters(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName string) (clusters map[string]bool, success bool) {

	request := threads.ProvisionerRequest{
		Action:     threads.ProvisionerGetClusters,
		Source:     threads.Http,
		ModuleName: moduleName,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).(map[string]bool), true
}

func GetSupervisors(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string) (map[uint64]supervisor.Status, bool) {

	request := threads.ProvisionerRequest{
		Action:      threads.ProvisionerGetSupervisors,
		Source:      threads.Http,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, timeout := utils.SendAndWait(responseTable, request.Nonce, GetConfigInstance().MaxWaitForResponse)
	if timeout {
		return nil, false
	}
	provisionerResponse := (data).(threads.ProvisionerResponse)

	return (provisionerResponse.Data).(map[uint64]supervisor.Status), true
}

func GetSupervisor(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string, supervisorId uint64) (supervisorInstance *supervisor.Supervisor, success bool) {

	request := threads.ProvisionerRequest{
		Action:      threads.ProvisionerGetSupervisor,
		Source:      threads.Http,
		ModuleName:  moduleName,
		ClusterName: clusterName,
		Metadata: threads.ProvisionerMetadata{
			SupervisorId: supervisorId,
		},
		Nonce: rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce, GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).(*supervisor.Supervisor), true
}

func FindStatistics(pipe chan<- threads.DatabaseRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (entries []database.Statistic, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.SupervisorStatistic,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(responseTable, databaseRequest.Nonce, GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	return (databaseResponse.Data).([]database.Statistic), true
}

func ShutdownCore(pipe chan<- threads.InterruptEvent) (response []byte, success bool) {
	pipe <- threads.Shutdown
	return nil, true
}

func PingNodeChannels(logger *utils.Logger,
	databasePipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	provisionerPipe chan<- threads.ProvisionerRequest, provisionerResponseTable *utils.ResponseTable) (success bool) {

	databasePingRequest := threads.DatabaseRequest{
		Action: threads.DatabaseUpperPing,
		Nonce:  rand.Uint32(),
	}
	databasePipe <- databasePingRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databasePingRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	if !databaseResponse.Success {
		return false
	}

	if GetConfigInstance().Debug {
		logger.Println("received ping over C2")
	}

	provisionerPingRequest := threads.ProvisionerRequest{
		Action: threads.ProvisionerLowerPing,
		Source: threads.Http,
		Nonce:  rand.Uint32(),
	}
	provisionerPipe <- provisionerPingRequest

	data2, didTimeout2 := utils.SendAndWait(provisionerResponseTable, provisionerPingRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout2 {
		return false
	}

	provisionerResponse := (data2).(threads.ProvisionerResponse)
	if !provisionerResponse.Success {
		return false
	}

	if GetConfigInstance().Debug {
		logger.Println("received ping over C6")
	}

	return true
}

func GetModules(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable) (success bool, modules []*provisioner.ModuleWrapper) {

	request := threads.ProvisionerRequest{
		Action: threads.ProvisionerGetModules,
		Source: threads.Http,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]*provisioner.ModuleWrapper)
}

func RegisterModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	modulePath string) (success bool, description string) {

	request := threads.ProvisionerRequest{
		Action: threads.ProvisionerModuleLoad,
		Source: threads.Http,
		Metadata: threads.ProvisionerMetadata{
			ModulePath: modulePath,
		},
		Nonce: rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, "timeout"
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)
	return provisionerResponse.Success, provisionerResponse.Description
}

func DeleteModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
	moduleName string) (success bool, description string) {

	request := threads.ProvisionerRequest{
		Action:     threads.ProvisionerModuleDelete,
		Source:     threads.Http,
		ModuleName: moduleName,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, "timeout"
	}

	provisionerResponse := (data).(threads.ProvisionerResponse)

	return provisionerResponse.Success, provisionerResponse.Description
}

func ToggleDebugMode(logger *utils.Logger) (description string) {

	cfg := GetConfigInstance()
	cfg.Debug = !cfg.Debug

	if cfg.Debug {
		description = "debug mode activated"
		logger.Println("remote change: debug mode ON")
	} else {
		description = "debug mode disabled"
		logger.Println("remote change: debug mode OFF")
	}

	return description
}
