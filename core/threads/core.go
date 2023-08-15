package threads

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	provisioner_component "github.com/GabeCordo/etl/core/components/provisioner"
	"github.com/GabeCordo/etl/core/threads/cache"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/threads/database"
	"github.com/GabeCordo/etl/core/threads/http"
	"github.com/GabeCordo/etl/core/threads/messenger"
	"github.com/GabeCordo/etl/core/threads/provisioner"
	"github.com/GabeCordo/etl/core/utils"
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
	core.C5 = make(chan threads.ProvisionerRequest, 10)
	core.C6 = make(chan threads.ProvisionerResponse, 10)
	core.C7 = make(chan threads.DatabaseRequest, 10)
	core.C8 = make(chan threads.DatabaseResponse, 10)
	core.C9 = make(chan threads.CacheRequest, 10)
	core.C10 = make(chan threads.CacheResponse, 10)
	core.C11 = make(chan threads.MessengerRequest, 10)
	core.interrupt = make(chan threads.InterruptEvent, 10)

	/* load the cfg in for the first time */
	if cfg := common.GetConfigInstance(configPath); cfg == nil {
		log.Panic("could not create cfg")
	}

	httpLogger, err := utils.NewLogger(utils.Http, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.HttpThread, err = http.NewThread(httpLogger, core.interrupt, core.C1, core.C2, core.C5, core.C6)
	if err != nil {
		return nil, err
	}

	provisionerLogger, err := utils.NewLogger(utils.Provisioner, &common.GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.ProvisionerThread, err = provisioner.NewThread(provisionerLogger, core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C9, core.C10, core.C11)
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
	core.DatabaseThread, err = database.NewThread(databaseLogger, DefaultConfigsFolder, DefaultStatisticsFolder,
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

	// needed in-case the supervisor or http threads need to populate Data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("Database Thread Started")
	}

	// if we chain requests, we should have a way to save that Data for re-use
	// FIX: the cache should start up before the provisioner in case the provisioner
	//		has stream processes that need to start using it.
	core.CacheThread.Setup()
	go core.CacheThread.Start()
	if common.GetConfigInstance().Debug {
		core.logger.Println("Cache Thread Started")
	}

	// we need a way to provision clusters if we are receiving threads before we can
	core.ProvisionerThread.Setup()
	go core.ProvisionerThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("Provisioner Thread Started")
	}

	// the gateway to the frontend cluster should be the last startup
	core.HttpThread.Setup()
	go core.HttpThread.Start() // event loop
	if common.GetConfigInstance().Debug {
		core.logger.Println("HTTP API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			common.GetConfigInstance().Net.Host, common.GetConfigInstance().Net.Port)
	}

	go core.repl()

	// monitor system calls being sent to the process, if the etl is being
	// run on a local machine, the developer might attempt to kill the process with SIGINT
	// requiring us to cleanly close the application without risking the loss of Data
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

	// close the gateway, stop new threads from flooding into the servers
	core.HttpThread.Teardown()

	core.logger.SetColour(utils.Red)

	if common.GetConfigInstance().Debug {
		core.logger.Println("http shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	core.ProvisionerThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("provisioner shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the Data is useless, shutdown
	core.CacheThread.Teardown()

	if common.GetConfigInstance().Debug {
		core.logger.Println("cache shutdown")
	}

	// the supervisor might need to store Data while finishing, close after
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

func (core *Core) Cluster(identifier string, mode cluster.EtlMode, implementation cluster.Cluster, configs ...cluster.Config) {

	p := provisioner.GetProvisionerInstance()
	defaultModule, _ := p.GetModule(provisioner_component.DefaultFrameworkModule) // the default threads module should always be found

	clusterWrapper, _ := defaultModule.AddCluster(identifier, mode, implementation)
	if common.GetConfigInstance().MountByDefault {
		clusterWrapper.Mount()
	} else {
		clusterWrapper.UnMount()
	}

	clusterImplementation := clusterWrapper.GetClusterImplementation()
	if helperImplementation, ok := (clusterImplementation).(cluster.Helper); ok {
		helper, _ := provisioner.NewHelper(provisioner_component.DefaultFrameworkModule, clusterWrapper.Identifier,
			core.C9, core.C11)
		helperImplementation.SetHelper(helper)
	}

	d := database.GetConfigDatabaseInstance()

	for _, config := range configs {
		d.Create(provisioner_component.DefaultFrameworkModule, identifier, config)
	}
}

func (core *Core) Module(path string) (success bool, description string) {

	success, description = common.RegisterModule(core.HttpThread.C5, core.HttpThread.ProvisionerResponseTable, path)
	return success, description
}
