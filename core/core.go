package core

import (
	"fmt"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"os"
	"os/signal"
	"syscall"
)

const (
	Version string = "v0.2.4-alpha"
)

func (core *Core) Banner() {
	fmt.Println("   ___    _____    _")
	fmt.Println("  | __|  |_   _|  | |")
	fmt.Println("  | _|     | |    | |__")
	fmt.Println("  |___|   _|_|_   |____|\taka. 'mango'")
	fmt.Println("_|\"\"\"\"\"|_|\"\"\"\"\"|_|\"\"\"\"\"|")
	fmt.Println("\"`-0-0-'\"`-0-0-'\"`-0-0-'")
	fmt.Println("[+] " + logging.Purple + "Extract Transform Load Framework " + logging.Reset + Version)
	fmt.Println("[+]" + logging.Purple + " by Gabriel Cordovado 2022-23" + logging.Reset)
	fmt.Println()
}

func (core *Core) Run() {

	core.Banner()

	core.logger.SetColour(logging.Purple)

	if GetConfigInstance().Debug {
		core.logger.Println("debug mode ON")
	} else {
		core.logger.Println("debug mode OFF")
	}

	// needed in-case the proceeding threads need logging or email capabilities during startup
	core.MessengerThread.Setup()
	go core.MessengerThread.Start() // event loop
	if core.config.Debug {
		core.logger.Println("Messenger Thread Started")
	}

	// needed in-case the supervisor or http_client threads need to populate data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if core.config.Debug {
		core.logger.Println("Database Thread Started")
	}

	// if we chain requests, we should have a way to save that data for re-use
	// FIX: the cache should start up before the provisioner in case the provisioner
	//		has stream processes that need to start using it.
	core.CacheThread.Setup()
	go core.CacheThread.Start()
	if core.config.Debug {
		core.logger.Println("Cache Thread Started")
	}

	core.SupervisorThread.Setup()
	go core.SupervisorThread.Start()
	if core.config.Debug {
		core.logger.Println("Supervisor Thread Started")
	}

	core.ProcessorThread.Setup()
	go core.ProcessorThread.Start()
	if core.config.Debug {
		core.logger.Println("Processor Thread Started")
	}

	core.SchedulerThread.Setup()
	go core.SchedulerThread.Start()
	if core.config.Debug {
		core.logger.Println("Scheduler Thread Starting")
		core.SchedulerThread.Scheduler.Print()
	}

	core.HttpProcessorThread.Setup()
	go core.HttpProcessorThread.Start()
	if core.config.Debug {
		core.logger.Println("HTTP Processor API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			core.config.Net.Processor.Host, core.config.Net.Processor.Port)
	}

	// the gateway to the frontend cluster should be the last startup
	core.HttpClientThread.Setup()
	go core.HttpClientThread.Start() // event loop
	if core.config.Debug {
		core.logger.Println("HTTP Client API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			core.config.Net.Client.Host, core.config.Net.Client.Port)
	}

	// HOTFIX: 3 - weird output for docker
	// bug: on docker having the REPL enabled causes the @etl prefix to
	// 		be spammed. Issue with docker giving the program the impression
	//		it is being fed empty lines?
	if core.config.EnableRepl {
		go core.repl()
	}

	// monitor system calls being sent to the process, if the etl is being
	// run on a local machine, the developer might attempt to kill the process with SIGINT
	// requiring us to cleanly close the application without risking the loss of data
	// ---
	// an interrupt can be sent by any thread that has access to the channel if an
	// error or end-state has been reached by the application
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		core.interrupt <- common.Panic
	case interrupt := <-core.interrupt:
		switch interrupt {
		case common.Panic:
			core.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			core.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	core.logger.SetColour(logging.Red)

	// close the gateway, stop new threads from flooding into the servers
	core.HttpClientThread.Teardown()

	if core.config.Debug {
		core.logger.Println("http_client shutdown")
	}

	core.HttpProcessorThread.Teardown()

	if core.config.Debug {
		core.logger.Println("http_processor shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	//core.ProvisionerThread.Teardown()
	//
	//if common.GetConfigInstance().Debug {
	//	core.logger.Println("provisioner shutdown")
	//}

	core.SchedulerThread.Teardown()

	if core.config.Debug {
		core.logger.Println("scheduler shutdown")
	}

	core.ProcessorThread.Teardown()

	if core.config.Debug {
		core.logger.Println("processor shutdown")
	}

	core.SupervisorThread.Teardown()

	if core.config.Debug {
		core.logger.Println("supervisor shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the data is useless, shutdown
	core.CacheThread.Teardown()

	if core.config.Debug {
		core.logger.Println("cache shutdown")
	}

	// the supervisor might need to store data while finishing, close after
	core.DatabaseThread.Teardown()

	if core.config.Debug {
		core.logger.Println("database shutdown")
	}

	// the preceding threads might need to log, or send alerts of failure during shutdown
	core.MessengerThread.Teardown()

	if core.config.Debug {
		core.logger.Println("messenger shutdown")
	}
}

// TODO : no longer support on the core
//func (core *Core) Cluster(identifier string, mode cluster.EtlMode, implementation cluster.Cluster, configs ...cluster.config) {
//
//	p := provisioner.GetProvisionerInstance()
//	defaultModule, _ := p.GetModule(provisioner_component.DefaultFrameworkModule) // the default threads module should always be found
//
//	clusterWrapper, _ := defaultModule.AddCluster(identifier, mode, implementation)
//	if common.GetConfigInstance().MountByDefault {
//		clusterWrapper.Mount()
//	} else {
//		clusterWrapper.UnMount()
//	}
//
//	clusterImplementation := clusterWrapper.GetClusterImplementation()
//	if helperImplementation, ok := (clusterImplementation).(cluster.Helper); ok {
//		helper, _ := provisioner.NewHelper(provisioner_component.DefaultFrameworkModule, clusterWrapper.Identifier,
//			core.C9, core.C17)
//		helperImplementation.SetHelper(helper)
//	}
//
//	d := database.GetConfigDatabaseInstance()
//
//	for _, config := range configs {
//		d.Create(provisioner_component.DefaultFrameworkModule, identifier, config)
//	}
//}

// TODO : no longer support on the core
//func (core *Core) Module(path string) (success bool, description string) {
//
//	success, description = common.RegisterModule(core.HttpClientThread.C5, core.HttpClientThread.ProcessorResponseTable, path)
//	return success, description
//}
