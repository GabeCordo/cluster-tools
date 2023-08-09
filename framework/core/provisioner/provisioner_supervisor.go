package provisioner

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/supervisor"
	"github.com/GabeCordo/etl/framework/core/common"
	"github.com/GabeCordo/etl/framework/utils"
	"time"
)

func (provisionerThread *Thread) ProcessGetSupervisors(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "module unknown"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "cluster unknown"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	supervisorStates := make(map[uint64]supervisor.Status)

	for _, supervisorInst := range clusterWrapper.FindSupervisors() {
		supervisorStates[supervisorInst.Id] = supervisorInst.State
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce,
		Data: supervisorStates}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessGetSupervisor(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "module unknown"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "cluster unknown"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	supervisorInstance, found := clusterWrapper.FindSupervisor(request.Metadata.SupervisorId)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "supervisor unknown"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce,
		Data: supervisorInstance}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessProvisionRequest(request *threads.ProvisionerRequest) {

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)

	if !found {
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision clusters from a mounted module
	// - if a module is unmounted, it is not meant to be operational
	if !moduleWrapper.IsMounted() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; it's module was not mounted\n", utils.Green, request.ModuleName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

	if !found {
		provisionerThread.logger.Warnf("%s[%s]%s Cluster does not exist\n", utils.Green, request.ClusterName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision mounted etl processes
	// - if a cluster is unmounted, even if the module is mounted, it is not meant to be operational
	if !clusterWrapper.IsMounted() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; cluster was not mounted\n", utils.Green, request.ClusterName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	// an operator shall only provision batch etl processes
	// - stream processes are meant to be run by the system when mounted or unmounted
	if (request.Source == threads.Http) && clusterWrapper.IsStream() {
		provisionerThread.logger.Warnf("%s[%s]%s Could not provision cluster; it's a stream process\n", utils.Green, request.ModuleName, utils.Reset)
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	provisionerThread.logger.Printf("%s[%s]%s Provisioning cluster in module %s\n", utils.Green, request.ClusterName, utils.Reset, request.ModuleName)

	// if the operator does not specify a common to use, the system shall use the cluster identifier name
	// to find a default common that should be located in the database thread
	if request.Metadata.ConfigName == "" {
		request.Metadata.ConfigName = request.ClusterName
	}

	cnf, configFound := common.GetConfigFromDatabase(provisionerThread.C7, provisionerThread.databaseResponseTable, request.ModuleName, request.Metadata.ConfigName)
	if !configFound {
		// the common was either never created or deleted from the database.
		// INSTEAD of continuing, the node should inform the user that the client cannot use the common they want
		response := &threads.ProvisionerResponse{Success: false, Description: "common not found", Nonce: request.Nonce}
		provisionerThread.Respond(request.Source, response)
		provisionerThread.requestWg.Done()
		return
	}

	var supervisorInstance *supervisor.Supervisor
	if configFound {
		provisionerThread.logger.Printf("%s[%s]%s Initializing cluster supervisor from common\n", utils.Green, request.ClusterName, utils.Reset)
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
	provisionerThread.Respond(request.Source, response)

	provisionerThread.logger.Printf("%s[%s]%s Cluster Running\n", utils.Green, request.ClusterName, utils.Reset)

	go func() {

		// block until the supervisor completes
		response := supervisorInstance.Start()

		// don't send the statistics of the cluster to the database unless an Identifier has been
		// given to the cluster for grouping purposes
		if len(supervisorInstance.Config.Identifier) != 0 {
			// saves statistics to the database thread
			dbRequest := &threads.DatabaseRequest{
				Action:  threads.DatabaseStore,
				Origin:  threads.Provisioner,
				Cluster: supervisorInstance.Config.Identifier,
				Data:    response,
			}
			provisionerThread.Request(threads.Database, dbRequest)

			// sends a completion message to the messenger thread to write to a log file or send an email regarding completion
			msgRequest := &threads.MessengerRequest{
				Action:  threads.MessengerClose,
				Cluster: supervisorInstance.Config.Identifier,
			}
			provisionerThread.Request(threads.Messenger, msgRequest)

			// provide the console with output indicating that the cluster has completed
			// we already provide output when a cluster is provisioned, so it completes the state
			if common.GetConfigInstance().Debug {
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
