package core

import (
	"ETLFramework/cluster"
	"log"
)

var supervisor *cluster.Supervisor

func GetSupervisorInstance() *cluster.Supervisor {

	if supervisor == nil {
		supervisor = cluster.NewSupervisor()
	}

	return supervisor
}

func (supervisorThread *SupervisorThread) Setup() {
	supervisorThread.accepting = true
	GetSupervisorInstance() // create the supervisor if it doesn't exist
}

func (supervisorThread *SupervisorThread) Start() {

	go func() {
		// request coming from http_server
		for request := range supervisorThread.C5 {
			if !supervisorThread.accepting {
				break
			}
			supervisorThread.wg.Add(1)
			supervisorThread.ProcessIncomingRequests(request)
		}

		supervisorThread.wg.Wait()
	}()
	go func() {
		for response := range supervisorThread.C8 {
			if !supervisorThread.accepting {
				break
			}
			supervisorThread.ProcessesIncomingResponses(response)
		}

		supervisorThread.wg.Wait()
	}()

	supervisorThread.wg.Wait()
}

func (supervisorThread *SupervisorThread) ProcessIncomingRequests(request SupervisorRequest) {
	switch request.Action {
	case Provision:
		{
			log.Printf("Provisioning cluster {%s}", request.Cluster)

			clstr, cnfg, ok := supervisor.Function(request.Cluster)
			if !ok {
				log.Println("there is a corrupted cluster in the supervisor")
				break
			}

			var m *cluster.Monitor
			if cnfg == nil {
				m = cluster.NewMonitor(*clstr)
			} else {
				m = cluster.NewCustomMonitor(*clstr, cnfg)
			}
			go func() {
				response := m.Start()

				// don't send the statistics of the cluster to the database unless an identifier has been
				// given to the cluster for grouping purposes
				if len(m.Config.Identifier) != 0 {
					request := DatabaseRequest{Action: Store, Origin: Supervisor, Cluster: m.Config.Identifier, Data: response}
					supervisorThread.C7 <- request
				}
				supervisorThread.wg.Done()
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

func (supervisorThread *SupervisorThread) ProcessesIncomingResponses(response DatabaseResponse) {

}

func (supervisorThread *SupervisorThread) Teardown() {
	supervisorThread.accepting = false

	supervisorThread.wg.Wait()
}
