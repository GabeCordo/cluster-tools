package provisioner

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"math/rand"
)

func (provisionerThread *Thread) Setup() {

	provisionerThread.accepting = true

	// initialize a provisioner instance with a common module
	GetProvisionerInstance()
}

func (provisionerThread *Thread) Start() {

	provisionerThread.listenersWg.Add(1)

	// temporary go-routine
	go func() {
		for _, moduleWrapper := range GetProvisionerInstance().GetModules() {

			for _, clusterWrapper := range moduleWrapper.GetClusters() {

				if clusterWrapper.IsMounted() && clusterWrapper.IsStream() {

					if !moduleWrapper.IsMounted() {
						moduleWrapper.Mount()
					}

					provisionerThread.logger.Printf("%s[%s]%s mount cluster \n", utils.Green, clusterWrapper.Identifier, utils.Reset)

					provisionerThread.C5 <- threads.ProvisionerRequest{
						Action:      threads.ProvisionerProvision,
						Source:      threads.Provisioner,
						ModuleName:  moduleWrapper.Identifier,
						ClusterName: clusterWrapper.Identifier,
						Metadata:    threads.ProvisionerMetadata{},
						Nonce:       rand.Uint32(),
					}
				}
			}
		}
	}()

	go func() {
		// request coming from http_server
		for request := range provisionerThread.C5 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.requestWg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessIncomingRequests(&request)
		}

		provisionerThread.listenersWg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C8 {
			if !provisionerThread.accepting {
				break
			}

			// if this doesn't spawn its own thread we will be left waiting
			provisionerThread.ProcessesIncomingDatabaseResponses(response)
		}

		provisionerThread.listenersWg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C10 {
			if !provisionerThread.accepting {
				break
			}

			// if this doesn't spawn its own thread we can be left waiting
			provisionerThread.ProcessIncomingCacheResponses(response)
		}

		provisionerThread.listenersWg.Wait()
	}()

	provisionerThread.listenersWg.Wait()
	provisionerThread.requestWg.Wait()
}

func (provisionerThread *Thread) ProcessIncomingRequests(request *threads.ProvisionerRequest) {

	switch request.Action {
	case threads.ProvisionerLowerPing:
		provisionerThread.ProcessPingProvisionerRequest(request)
	case threads.ProvisionerMount:
		provisionerThread.ProcessMountRequest(request)
	case threads.ProvisionerUnMount:
		provisionerThread.ProcessUnMountRequest(request)
	case threads.ProvisionerProvision:
		provisionerThread.ProcessProvisionRequest(request)
	case threads.ProvisionerGetModules:
		provisionerThread.ProcessGetModules(request)
	case threads.ProvisionerModuleLoad:
		provisionerThread.ProcessAddModule(request)
	case threads.ProvisionerModuleDelete:
		provisionerThread.ProcessDeleteModule(request)
	case threads.ProvisionerGetClusters:
		provisionerThread.ProcessGetClusters(request)
	case threads.ProvisionerGetSupervisors:
		provisionerThread.ProcessGetSupervisors(request)
	case threads.ProvisionerGetSupervisor:
		provisionerThread.ProcessGetSupervisor(request)
	default:
		provisionerThread.logger.Println("got unknown request action")
	}
}

func (provisionerThread *Thread) Request(module threads.Module, request any) (success bool) {

	success = true

	switch module {
	case threads.Database:
		provisionerThread.C7 <- *(request).(*threads.DatabaseRequest)
	case threads.Cache:
		provisionerThread.C9 <- *(request).(*threads.CacheRequest)
	case threads.Messenger:
		provisionerThread.C11 <- *(request).(*threads.MessengerRequest)
	default:
		success = false
	}

	return success
}

func (provisionerThread *Thread) Respond(module threads.Module, response any) (success bool) {

	success = true
	switch module {
	case threads.Http:
		provisionerThread.C6 <- *(response).(*threads.ProvisionerResponse)
	default:
		success = false
	}

	return true
}

func (provisionerThread *Thread) ProcessPingProvisionerRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	if common.GetConfigInstance().Debug {
		provisionerThread.logger.Println("received ping over C5")
	}

	databaseRequest := &threads.DatabaseRequest{
		Action: threads.DatabaseLowerPing,
		Nonce:  rand.Uint32(),
		Data:   make(map[string]string),
	}
	provisionerThread.Request(threads.Database, databaseRequest)

	rawResponse, databasePingTimeout := utils.SendAndWait(provisionerThread.databaseResponseTable, databaseRequest.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if databasePingTimeout {
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		return
	}

	databaseResponse := (rawResponse).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		return
	}

	if common.GetConfigInstance().Debug {
		provisionerThread.logger.Println("received ping over C8")
	}

	cacheRequest := &threads.CacheRequest{Action: threads.CacheLowerPing, Nonce: rand.Uint32()}
	provisionerThread.Request(threads.Cache, cacheRequest)

	rawResponse, cachePingTimeout := utils.SendAndWait(provisionerThread.cacheResponseTable, cacheRequest.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if cachePingTimeout {
		provisionerThread.logger.Alertln("failed to receive ping over C10")
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		return
	}

	cacheResponse := (rawResponse).(threads.CacheResponse)

	if !cacheResponse.Success {
		provisionerThread.logger.Alertln("failed to receive ping over C10")
		response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: false}
		provisionerThread.Respond(request.Source, response)
		return
	}

	if common.GetConfigInstance().Debug {
		provisionerThread.logger.Println("[etl_provisioner] received ping over C10")
	}

	response := &threads.ProvisionerResponse{Nonce: request.Nonce, Success: true}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessesIncomingDatabaseResponses(response threads.DatabaseResponse) {
	provisionerThread.databaseResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *Thread) ProcessIncomingCacheResponses(response threads.CacheResponse) {
	provisionerThread.cacheResponseTable.Write(response.Nonce, response)
}

func (provisionerThread *Thread) Teardown() {
	provisionerThread.accepting = false

	provisionerThread.requestWg.Wait()
}
