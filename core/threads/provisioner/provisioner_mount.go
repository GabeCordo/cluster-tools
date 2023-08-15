package provisioner

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"math/rand"
)

func (provisionerThread *Thread) ProcessMountRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.Respond(request.Source, response)
		return
	}

	if !moduleWrapper.IsMounted() {
		moduleWrapper.Mount()

		if common.GetConfigInstance().Debug {
			provisionerThread.logger.Printf("%s[%s]%s Mounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			provisionerThread.logger.Printf("%s[%s]%s could not find cluster\n", utils.Green, request.ModuleName, utils.Reset)
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.Respond(request.Source, response)
			return
		}

		if !clusterWrapper.IsMounted() {
			clusterWrapper.Mount()

			if common.GetConfigInstance().Debug {
				provisionerThread.logger.Printf("%s[%s]%s Mounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}

			if clusterWrapper.IsStream() {
				provisionerThread.C5 <- threads.ProvisionerRequest{
					Action:      threads.ProvisionerProvision,
					Source:      threads.Provisioner,
					ModuleName:  request.ModuleName,
					ClusterName: request.ClusterName,
					Metadata:    threads.ProvisionerMetadata{},
					Nonce:       rand.Uint32(),
				}
			}
		}
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.Respond(request.Source, response)
}

func (provisionerThread *Thread) ProcessUnMountRequest(request *threads.ProvisionerRequest) {

	defer provisionerThread.requestWg.Done()

	moduleWrapper, found := GetProvisionerInstance().GetModule(request.ModuleName)
	if !found {
		response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
		provisionerThread.Respond(request.Source, response)
		return
	}

	if moduleWrapper.IsMounted() {
		moduleWrapper.UnMount()

		if common.GetConfigInstance().Debug {
			provisionerThread.logger.Printf("%s[%s]%s UnMounted module\n", utils.Green, request.ModuleName, utils.Reset)
		}
	}

	if request.ClusterName != "" {
		clusterWrapper, found := moduleWrapper.GetCluster(request.ClusterName)

		if !found {
			response := &threads.ProvisionerResponse{Success: false, Nonce: request.Nonce}
			provisionerThread.Respond(request.Source, response)
			return
		}

		if clusterWrapper.IsMounted() {
			clusterWrapper.UnMount()

			if common.GetConfigInstance().Debug {
				provisionerThread.logger.Printf("%s[%s]%s UnMounted cluster\n", utils.Green, request.ClusterName, utils.Reset)
			}

			if clusterWrapper.Mode == cluster.Stream {
				fmt.Println("suspend supervisor")
				clusterWrapper.SuspendSupervisors()
			}
		}
	}

	response := &threads.ProvisionerResponse{Success: true, Nonce: request.Nonce}
	provisionerThread.Respond(request.Source, response)
}
