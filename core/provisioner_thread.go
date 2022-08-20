package core

import (
	"ETLFramework/cluster"
	"log"
	"time"
)

var provisioner *cluster.Provisioner

func GetProvisionerInstance() *cluster.Provisioner {

	if provisioner == nil {
		provisioner = cluster.NewProvisioner()
	}

	return provisioner
}

func (provisionerThread *ProvisionerThread) Setup() {
	provisionerThread.accepting = true
	GetProvisionerInstance() // create the supervisor if it doesn't exist
}

func (provisionerThread *ProvisionerThread) Start() {
	go func() {
		// request coming from http_server
		for request := range provisionerThread.C5 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.wg.Add(1)
			provisionerThread.ProcessIncomingRequests(request)
		}

		provisionerThread.wg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C8 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.ProcessesIncomingResponses(response)
		}

		provisionerThread.wg.Wait()
	}()

	provisionerThread.wg.Wait()
}

func (provisionerThread *ProvisionerThread) ProcessIncomingRequests(request ProvisionerRequest) {
	switch request.Action {
	case Provision:
		{
			log.Printf("Provisioning cluster {%s}", request.Cluster)

			clstr, cnfg, ok := provisioner.Function(request.Cluster)
			if !ok {
				log.Println("there is a corrupted cluster in the supervisor")
				break
			}

			var m *cluster.Supervisor
			if cnfg == nil {
				m = cluster.NewSupervisor(*clstr)
			} else {
				m = cluster.NewCustomSupervisor(*clstr, cnfg)
			}
			go func() {
				var response *cluster.Response

				c := make(chan struct{})
				go func() {
					defer close(c)
					response = m.Start()
				}()

				go func() {
					defer close(c)
					for provisionerThread.accepting {
						// block
					}
					<-time.After(25 * time.Second)
				}()

				<-c

				// don't send the statistics of the cluster to the database unless an identifier has been
				// given to the cluster for grouping purposes
				if len(m.Config.Identifier) != 0 {
					request := DatabaseRequest{Action: Store, Origin: Provisioner, Cluster: m.Config.Identifier, Data: response}
					provisionerThread.C7 <- request
				}
				provisionerThread.wg.Done()
			}()
			break
		}
	case Teardown:
		{
			// TODO - not implemented
			break
		}
	default:
		{

		}
	}
}

func (provisionerThread *ProvisionerThread) ProcessesIncomingResponses(response DatabaseResponse) {

}

func (provisionerThread *ProvisionerThread) Teardown() {
	provisionerThread.accepting = false

	provisionerThread.wg.Wait()
}
