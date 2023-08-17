package threads

import (
	"fmt"
	core_i "github.com/GabeCordo/etl-light/core"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/threads/cache"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/threads/database"
	"github.com/GabeCordo/etl/core/threads/http_client"
	"github.com/GabeCordo/etl/core/threads/http_processor"
	"github.com/GabeCordo/etl/core/threads/messenger"
	"github.com/GabeCordo/etl/core/threads/processor"
	"github.com/GabeCordo/etl/core/threads/supervisor"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	Version string = "v0.1.9-alpha"
)

func NewCore(configPath string) (*Core, error) {
	core := new(Core)

	core.C1 = make(chan threads.DatabaseRequest, 10)
	core.C2 = make(chan threads.DatabaseResponse, 10)
	core.C3 = make(chan threads.MessengerRequest, 10)
	core.C4 = make(chan threads.MessengerResponse, 10)
	core.C5 = make(chan common.ProcessorRequest, 10)
	core.C6 = make(chan common.ProcessorResponse, 10)
	core.C7 = make(chan threads.DatabaseRequest, 10)
	core.C8 = make(chan threads.DatabaseResponse, 10)
	core.C9 = make(chan threads.CacheRequest, 10)
	core.C10 = make(chan threads.CacheResponse, 10)
	core.C11 = make(chan threads.MessengerRequest, 10)
	core.C12 = make(chan common.ProcessorRequest, 10)
	core.C13 = make(chan common.ProcessorResponse, 10)
	core.C14 = make(chan common.SupervisorRequest, 10)
	core.C15 = make(chan common.SupervisorResponse, 10)
	core.interrupt = make(chan threads.InterruptEvent, 10)

	/* load the cfg in for the first time */
	if cfg := common.GetConfigInstance(configPath); cfg == nil {
		log.Panic("could not create cfg")
	}

	httpLogger, err := utils.NewLogger(utils.HttpClient, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.HttpClientThread, err = http_client.NewThread(httpLogger,
		core.interrupt, core.C1, core.C2, core.C5, core.C6)
	if err != nil {
		return nil, err
	}

	httpProcessorLogger, err := utils.NewLogger(utils.HttpProcessor, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.HttpProcessorThread, err = http_processor.NewThread(httpProcessorLogger,
		core.interrupt, core.C12, core.C13)
	if err != nil {
		return nil, err
	}

	processorLogger, err := utils.NewLogger(utils.Processor, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.ProcessorThread, err = processor.NewThread(processorLogger,
		core.interrupt, core.C5, core.C6, core.C12, core.C13, core.C14, core.C15)
	if err != nil {
		return nil, err
	}

	supervisorLogger, err := utils.NewLogger(utils.Supervisor, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.SupervisorThread, err = supervisor.NewThread(supervisorLogger,
		core.interrupt, core.C14, core.C15, core.C7, core.C8, core.C9, core.C10, core.C11)
	if err != nil {
		return nil, err
	}

	messengerLogger, err := utils.NewLogger(utils.Messenger, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.MessengerThread, err = messenger.NewThread(messengerLogger, core.interrupt, core.C3, core.C4, core.C11)
	if err != nil {
		return nil, err
	}

	databaseLogger, err := utils.NewLogger(utils.Database, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.DatabaseThread, err = database.NewThread(databaseLogger, core_i.DefaultConfigsFolder, core_i.DefaultStatisticsFolder,
		core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C7, core.C8)
	if err != nil {
		return nil, err
	}

	cacheLogger, err := utils.NewLogger(utils.Cache, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.CacheThread, err = cache.NewThread(cacheLogger, core.interrupt, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	coreLogger, err := utils.NewLogger(utils.Undefined, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.logger = coreLogger

	return core, nil
}

func (core *Core) Banner() {
	fmt.Println("   ___    _____    _")
	fmt.Println("  | __|  |_   _|  | |")
	fmt.Println("  | _|     | |    | |__")
	fmt.Println("  |___|   _|_|_   |____|")
	fmt.Println("_|\"\"\"\"\"|_|\"\"\"\"\"|_|\"\"\"\"\"|")
	fmt.Println("\"`-0-0-'\"`-0-0-'\"`-0-0-'")
	fmt.Println("[+] " + utils.Purple + "Extract Transform Load Framework " + utils.Reset + Version)
	fmt.Println("[+]" + utils.Purple + " by Gabriel Cordovado 2022-23" + utils.Reset)
	fmt.Println()
}

func (core *Core) Run() {

	core.Banner()

	core.logger.SetColour(utils.Purple)

	if common.GetConfigInstance().Debug {
		core.logger.Println("debug mode ON")
	} else {
		core.logger.Println("debug mode OFF")
	}

	// needed in-case the proceeding threads need logging or email capabilities during startup
	core.MessengerThread.Setup()
	go core.MessengerThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("Messenger Thread Started")
	}

	// needed in-case the supervisor or http_client threads need to populate data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("Database Thread Started")
	}

	// if we chain requests, we should have a way to save that data for re-use
	// FIX: the cache should start up before the provisioner in case the provisioner
	//		has stream processes that need to start using it.
	core.CacheThread.Setup()
	go core.CacheThread.Start()
	if common.GetConfigInstance().Debug {
		core.logger.Println("Cache Thread Started")
	}

	// we need a way to provision clusters if we are receiving threads before we can
	//core.ProvisionerThread.Setup()
	//go core.ProvisionerThread.Start() // event loop
	//if common.GetConfigInstance().Debug {
	//	core.logger.Println("Provisioner Thread Started")
	//}

	core.SupervisorThread.Setup()
	go core.SupervisorThread.Start()
	if common.GetConfigInstance().Debug {
		core.logger.Println("Supervisor Thread Started")
	}

	core.ProcessorThread.Setup()
	go core.ProcessorThread.Start()
	if common.GetConfigInstance().Debug {
		core.logger.Println("Processor Thread Started")
	}

	core.HttpProcessorThread.Setup()
	go core.HttpProcessorThread.Start()
	if common.GetConfigInstance().Debug {
		core.logger.Println("HTTP Processor API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			common.GetConfigInstance().Net.Processor.Host, common.GetConfigInstance().Net.Processor.Port)
	}

	// the gateway to the frontend cluster should be the last startup
	core.HttpClientThread.Setup()
	go core.HttpClientThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("HTTP Client API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			common.GetConfigInstance().Net.Client.Host, common.GetConfigInstance().Net.Client.Port)
	}

	go core.repl()

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
		core.interrupt <- threads.Panic
	case interrupt := <-core.interrupt:
		switch interrupt {
		case threads.Panic:
			core.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			core.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	core.logger.SetColour(utils.Red)

	// close the gateway, stop new threads from flooding into the servers
	core.HttpClientThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("http_client shutdown")
	}

	core.HttpProcessorThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("http_processor shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	//core.ProvisionerThread.Teardown()
	//
	//if common.GetConfigInstance().Debug {
	//	core.logger.Println("provisioner shutdown")
	//}

	core.ProcessorThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("processor shutdown")
	}

	core.SupervisorThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("supervisor shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the data is useless, shutdown
	core.CacheThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("cache shutdown")
	}

	// the supervisor might need to store data while finishing, close after
	core.DatabaseThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("database shutdown")
	}

	// the preceding threads might need to log, or send alerts of failure during shutdown
	core.MessengerThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("messenger shutdown")
	}
}

// TODO : no longer support on the core
//func (core *Core) Cluster(identifier string, mode cluster.EtlMode, implementation cluster.Cluster, configs ...cluster.Config) {
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
//			core.C9, core.C11)
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
