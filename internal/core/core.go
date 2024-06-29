package core

import (
	"github.com/GabeCordo/cluster-tools/internal/core/threads/cache"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/database"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/http_client"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/http_processor"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/messenger"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/scheduler"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/supervisor"
	"github.com/GabeCordo/toolchain/logging"
	"os"
	"os/signal"
	"syscall"
)

type Core struct {
	HttpClientThread    *http_client.Thread
	HttpProcessorThread *http_processor.Thread
	ProcessorThread     *processor.Thread
	SupervisorThread    *supervisor.Thread
	MessengerThread     *messenger.Thread
	DatabaseThread      *database.Thread
	CacheThread         *cache.Thread
	SchedulerThread     *scheduler.Thread

	C1        chan common.ThreadRequest  // DatabaseRequest
	C2        chan common.ThreadResponse // DatabaseResponse
	C3        chan common.ThreadRequest  // MessengerRequest
	C4        chan common.ThreadResponse // MessengerResponse
	C5        chan common.ThreadRequest  // ProcessorRequest
	C6        chan common.ThreadResponse // ProcessorResponse
	C7        chan common.ThreadRequest  // ProcessorRequest
	C8        chan common.ThreadResponse // ProcessorResponse
	C9        chan common.ThreadRequest  // CacheRequest
	C10       chan common.ThreadResponse // CacheResponse
	C11       chan common.ThreadRequest  // DatabaseRequest
	C12       chan common.ThreadResponse // DatabaseResponse
	C13       chan common.ThreadRequest  // SupervisorRequest
	C14       chan common.ThreadResponse // SupervisorResponse
	C15       chan common.ThreadRequest  // DatabaseRequest
	C16       chan common.ThreadResponse // DatabaseResponse
	C17       chan common.ThreadRequest  // MessengerRequest
	C18       chan common.ThreadRequest  // ProcessorRequest
	C19       chan common.ThreadResponse // ProcessorResponse
	C20       chan common.ThreadRequest  // SchedulerRequest
	C21       chan common.ThreadResponse // SchedulerResponse
	C22       chan common.ThreadRequest  // MessengerRequest
	C23       chan common.ThreadResponse // MessengerResponse
	C24       chan common.ThreadRequest  // CacheRequest
	C25       chan common.ThreadResponse // CacheResponse
	C26       chan common.ThreadRequest  // CacheRequest
	C27       chan common.ThreadResponse // CacheResponse
	interrupt chan common.InterruptEvent // InterruptEvent

	config *Config
	logger *logging.Logger
}

func New(configPath string) (*Core, error) {
	core := new(Core)

	core.interrupt = make(chan common.InterruptEvent, 10)
	core.C1 = make(chan common.ThreadRequest, 10)
	core.C2 = make(chan common.ThreadResponse, 10)
	core.C3 = make(chan common.ThreadRequest, 10)
	core.C4 = make(chan common.ThreadResponse, 10)
	core.C5 = make(chan common.ThreadRequest, 10)
	core.C6 = make(chan common.ThreadResponse, 10)
	core.C7 = make(chan common.ThreadRequest, 10)
	core.C8 = make(chan common.ThreadResponse, 10)
	core.C9 = make(chan common.ThreadRequest, 10)
	core.C10 = make(chan common.ThreadResponse, 10)
	core.C11 = make(chan common.ThreadRequest, 10)
	core.C12 = make(chan common.ThreadResponse, 10)
	core.C13 = make(chan common.ThreadRequest, 10)
	core.C14 = make(chan common.ThreadResponse, 10)
	core.C15 = make(chan common.ThreadRequest, 10)
	core.C16 = make(chan common.ThreadResponse, 10)
	core.C17 = make(chan common.ThreadRequest, 10)
	core.C18 = make(chan common.ThreadRequest, 10)
	core.C19 = make(chan common.ThreadResponse, 10)
	core.C20 = make(chan common.ThreadRequest, 10)
	core.C21 = make(chan common.ThreadResponse, 10)
	core.C22 = make(chan common.ThreadRequest, 10)
	core.C23 = make(chan common.ThreadResponse, 10)
	core.C24 = make(chan common.ThreadRequest, 10)
	core.C25 = make(chan common.ThreadResponse, 10)
	core.C26 = make(chan common.ThreadRequest, 10)
	core.C27 = make(chan common.ThreadResponse, 10)

	/* load the cfg in for the first time */
	core.config = GetConfigInstance(configPath)

	httpLogger, err := logging.NewLogger(HttpClient.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	httpConfig := &http_client.Config{}
	core.config.FillHttpClientConfig(httpConfig)
	core.HttpClientThread, err = http_client.New(httpConfig, httpLogger,
		core.interrupt, core.C1, core.C2, core.C5, core.C6, core.C20, core.C21, core.C22, core.C23, core.C24, core.C25)
	if err != nil {
		return nil, err
	}

	httpProcessorLogger, err := logging.NewLogger(HttpProcessor.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	httpProcessorConfig := &http_processor.Config{}
	core.config.FillHttpProcessorConfig(httpProcessorConfig)
	core.HttpProcessorThread, err = http_processor.New(httpProcessorConfig, httpProcessorLogger,
		core.interrupt, core.C7, core.C8, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Processor.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	processorConfig := &processor.Config{}
	core.config.FillProcessorConfig(processorConfig)
	core.ProcessorThread, err = processor.New(processorConfig, processorLogger,
		core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C11, core.C12, core.C13, core.C14, core.C18, core.C19)
	if err != nil {
		return nil, err
	}

	supervisorLogger, err := logging.NewLogger(Supervisor.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	supervisorConfig := &supervisor.Config{}
	core.config.FillSupervisorConfig(supervisorConfig)
	core.SupervisorThread, err = supervisor.NewThread(supervisorConfig, supervisorLogger,
		core.interrupt, core.C13, core.C14, core.C15, core.C16, core.C17)
	if err != nil {
		return nil, err
	}

	messengerLogger, err := logging.NewLogger(Messenger.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	messengerConfig := &messenger.Config{}
	core.config.FillMessengerConfig(messengerConfig)
	core.MessengerThread, err = messenger.New(messengerConfig, messengerLogger,
		core.interrupt, core.C3, core.C4, core.C17, core.C22, core.C23)
	if err != nil {
		return nil, err
	}

	databaseLogger, err := logging.NewLogger(Database.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	databaseConfig := &database.Config{}
	core.config.FillDatabaseConfig(databaseConfig)
	core.DatabaseThread, err = database.New(databaseConfig, databaseLogger,
		common.DefaultConfigsFolder, common.DefaultStatisticsFolder,
		core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C11, core.C12, core.C15, core.C16, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	cacheLogger, err := logging.NewLogger(Cache.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	cacheConfig := &cache.Config{}
	core.config.FillCacheConfig(cacheConfig)
	core.CacheThread, err = cache.New(cacheConfig, cacheLogger,
		core.interrupt, core.C9, core.C10, core.C24, core.C25)
	if err != nil {
		return nil, err
	}

	schedulerLogger, err := logging.NewLogger(Scheduler.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	schedulerConig := &scheduler.Config{}
	core.config.FillSchedulerConfig(schedulerConig)
	core.SchedulerThread, err = scheduler.New(schedulerConig, schedulerLogger,
		core.interrupt, core.C18, core.C19, core.C20, core.C21, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	coreLogger, err := logging.NewLogger(Undefined.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.logger = coreLogger

	return core, nil
}

const (
	Version string = "v0.2.4-alpha"
)

func (core *Core) Run() {

	core.banner()

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
