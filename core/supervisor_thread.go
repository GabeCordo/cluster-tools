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

func (supervisorThread *Supervisor) Setup() {
	supervisorThread.accepting = true
	GetSupervisorInstance() // create the supervisor if it doesn't exist
}

func (supervisorThread *Supervisor) Start() {
	var request SupervisorRequest
	for supervisorThread.accepting {
		request = <-supervisorThread.C5 // request coming from http_server

		switch request.Action {
		case Provision:
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
				m = cluster.NewCustomMonitor(*clstr, *cnfg)
			}
			go m.Start()
			break
		case Teardown:
			// TODO - not implemented
			break
		}
	}

	supervisorThread.wg.Wait()
}

func (supervisorThread *Supervisor) Teardown() {
	supervisorThread.accepting = false
}
