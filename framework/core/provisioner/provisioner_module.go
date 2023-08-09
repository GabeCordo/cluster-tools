package provisioner

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/module"
	"github.com/GabeCordo/etl/framework/core/common"
	"github.com/GabeCordo/etl/framework/utils"
	"math/rand"
	"reflect"
)

func (provisionerThread *Thread) ProcessGetModules(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()
	provisionerThread.logger.Println("getting modules in provisioner")

	provisioner := GetProvisionerInstance()
	modules := provisioner.GetModules()

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce, Data: modules}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessGetClusters(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerInstance := GetProvisionerInstance()

	clusters := make(map[string]bool, 0)

	moduleWrapper, found := provisionerInstance.GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce,
			Description: "module not found"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	mounts := moduleWrapper.GetClustersData()
	for identifier, isMounted := range mounts {
		clusters[identifier] = isMounted
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce,
		Data: clusters}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessAddModule(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerThread.logger.Printf("registering module at %s\n", request.Metadata.ModulePath)

	remoteModule, err := module.NewRemoteModule(request.Metadata.ModulePath)
	if err != nil {
		provisionerThread.logger.Alertln("cannot find remote module")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "cannot find remote module"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	moduleInstance, err := remoteModule.Get()
	if err != nil {
		provisionerThread.logger.Alertln("module built with older version")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "module built with older version"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	if err := GetProvisionerInstance().AddModule(moduleInstance); err != nil {
		provisionerThread.logger.Alertln("a module with that identifier already exists or is corrupt")
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "a module with that identifier already exists or is corrupt"}
		provisionerThread.Respond(request.Source, response)
		return
	}

	// note: the module wrapper should already be defined so there is no need to validate
	moduleWrapper, _ := GetProvisionerInstance().GetModule(moduleInstance.Config.Identifier)

	registeredClusters := moduleWrapper.GetClusters()

	// REGISTER ANY HELPERS TO CLUSTERS THAT HAVE DEFINED THEM WITHIN THE STRUCT
	for _, clusterWrapper := range registeredClusters {

		clusterImplementation := clusterWrapper.GetClusterImplementation()

		reflectedClusterImpl := reflect.ValueOf(clusterImplementation)

		helperField := reflect.Indirect(reflectedClusterImpl).FieldByName("Helper")
		if helperField.IsValid() && helperField.CanSet() {
			if helper, err := NewHelper(provisionerThread.C9, provisionerThread.C11); err == nil {
				helperField.Set(reflect.ValueOf(helper))
			}
		}
	}

	// get all the configs from the original list that were sucessfully created
	registeredExports := make([]module.Cluster, 0)
	for _, export := range moduleInstance.Config.Exports {
		if _, found := moduleWrapper.GetCluster(export.Cluster); found {
			registeredExports = append(registeredExports, export)
		}
	}

	registeredConfigs := make(map[string]cluster.Config)

	// REGISTER EACH CONFIG FROM THE MODULE FILE TO THE DATABASE THREAD
	for _, export := range registeredExports {

		cfg := cluster.Config{
			Identifier:                  export.Cluster,
			OnCrash:                     export.Config.OnCrash,
			OnLoad:                      export.Config.OnLoad,
			StartWithNLoadClusters:      export.Config.Static.LFunctions,
			StartWithNTransformClusters: export.Config.Static.TFunctions,
			ETChannelThreshold:          export.Config.Dynamic.TFunction.Threshold,
			ETChannelGrowthFactor:       export.Config.Dynamic.TFunction.GrowthFactor,
			TLChannelThreshold:          export.Config.Dynamic.LFunction.Threshold,
			TLChannelGrowthFactor:       export.Config.Dynamic.LFunction.GrowthFactor,
		}

		if !cfg.Valid() {
			cfg = cluster.DefaultConfig
			cfg.Identifier = export.Cluster
		}

		registeredConfigs[cfg.Identifier] = cfg

		databaseRequest := &threads.DatabaseRequest{
			Action:  threads.DatabaseStore,
			Nonce:   rand.Uint32(),
			Origin:  threads.Provisioner,
			Type:    threads.ClusterConfig,
			Module:  moduleInstance.Config.Identifier,
			Cluster: export.Cluster,
			Data:    cfg,
		}
		provisionerThread.Request(threads.Database, databaseRequest)

		data, didTimeout := utils.SendAndWait(provisionerThread.databaseResponseTable, databaseRequest.Nonce,
			common.GetConfigInstance().MaxWaitForResponse)

		if didTimeout {
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "could not save cfg"}
			provisionerThread.Respond(request.Source, response)
			return
		}

		response := (data).(threads.DatabaseResponse)

		if !response.Success {
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce, Description: "could not save cfg"}
			provisionerThread.Respond(request.Source, response)
			return
		}
	}

	provisionerThread.logger.Printf("dynamically loaded module %s\n", moduleInstance.Config.Identifier)

	// FOR EVERY CLUSTER WITH A MODE OF 'STREAM' PROVISION IT
	for _, export := range registeredExports {

		if !export.StaticMount {
			continue
		}

		provisionRequest := threads.ProvisionerRequest{
			Action:      threads.ProvisionerProvision,
			ModuleName:  moduleInstance.Config.Identifier,
			ClusterName: export.Cluster,
			Metadata: threads.ProvisionerMetadata{
				ConfigName: export.Cluster,
				Other:      make(map[string]string),
			},
			Source: threads.Provisioner,
			Nonce:  rand.Uint32(),
		}

		provisionerThread.C5 <- provisionRequest
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce, Description: "module registered"}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessDeleteModule(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	provisionerInstance := GetProvisionerInstance()

	response := &threads.ProvisionerResponse{Nonce: request.Nonce}
	if deleted, _, found := provisionerInstance.DeleteModule(request.ModuleName); found {
		response.Success = true
		if deleted {
			response.Description = "module deleted"

			databaseRequest := &threads.DatabaseRequest{
				Action: threads.DatabaseDelete,
				Type:   threads.ClusterModule,
				Module: request.ModuleName,
				Nonce:  rand.Uint32(),
			}
			provisionerThread.Request(threads.Database, databaseRequest)

			data, didTimeout := utils.SendAndWait(provisionerThread.databaseResponseTable, databaseRequest.Nonce,
				common.GetConfigInstance().MaxWaitForResponse)

			if didTimeout {
				response.Success = false
				response.Description = "could not delete clusters and statistics under a module"
			}

			databaseResponse := (data).(threads.DatabaseResponse)

			if !databaseResponse.Success {
				response.Success = false
				response.Description = "could not delete clusters and statistics under a module"
			}
		} else {
			response.Description = "module marked for deletion, a cluster is likely running right now, try later"
		}
	} else {
		response.Success = false
		response.Description = "module not found"
	}

	provisionerThread.Respond(request.Source, response)
}
