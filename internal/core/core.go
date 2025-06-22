package core

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/GabeCordo/Flock/internal/core/component/message/log"
	processorCmp "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/job"
	configDb "github.com/GabeCordo/Flock/internal/core/database/pipeline"
	supervisorDb "github.com/GabeCordo/Flock/internal/core/database/run"
	statisticDb "github.com/GabeCordo/Flock/internal/core/database/statistic"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/Flock/internal/core/thread/database"
	"github.com/GabeCordo/Flock/internal/core/thread/messenger"
	"github.com/GabeCordo/Flock/internal/core/thread/processor"
	restApi "github.com/GabeCordo/Flock/internal/core/thread/rest"
	"github.com/GabeCordo/Flock/internal/core/thread/runner"
	"github.com/GabeCordo/Flock/internal/core/thread/scheduler"
	"github.com/GabeCordo/Flock/internal/core/thread/socket"
	"github.com/GabeCordo/toolchain/logging"
)

type Core struct {
	RestThread      *restApi.Thread
	SocketThread    *socket.Thread
	ProcessorThread *processor.Thread
	RunnerThread    *runner.Thread
	MessengerThread *messenger.Thread
	DatabaseThread  *database.Thread
	SchedulerThread *scheduler.Thread

	C1        chan *thread.Request       // DatabaseRequest
	C2        chan *thread.Response      // DatabaseResponse
	C3        chan *thread.Request       // MessengerRequest
	C4        chan *thread.Response      // MessengerResponse
	C5        chan *thread.Request       // ProcessorRequest
	C6        chan *thread.Response      // ProcessorResponse
	C7        chan *thread.Request       // ProcessorRequest
	C8        chan *thread.Response      // ProcessorResponse
	C9        chan *thread.Request       // CacheRequest
	C10       chan *thread.Response      // CacheResponse
	C11       chan *thread.Request       // DatabaseRequest
	C12       chan *thread.Response      // DatabaseResponse
	C13       chan *thread.Request       // SupervisorRequest
	C14       chan *thread.Response      // SupervisorResponse
	C15       chan *thread.Request       // DatabaseRequest
	C16       chan *thread.Response      // DatabaseResponse
	C17       chan *thread.Request       // MessengerRequest
	C18       chan *thread.Request       // ProcessorRequest
	C19       chan *thread.Response      // ProcessorResponse
	C20       chan *thread.Request       // SchedulerRequest
	C21       chan *thread.Response      // SchedulerResponse
	C22       chan *thread.Request       // MessengerRequest
	C23       chan *thread.Response      // MessengerResponse
	C24       chan *thread.Request       // CacheRequest
	C25       chan *thread.Response      // CacheResponse
	C26       chan *thread.Request       // CacheRequest
	C27       chan *thread.Response      // CacheResponse
	interrupt chan thread.InterruptEvent // InterruptEvent

	config *Config
	logger *logging.Logger
}

func New(configPath string) (*Core, error) {
	core := new(Core)

	core.interrupt = make(chan thread.InterruptEvent, 10)
	core.C1 = make(chan *thread.Request, 10)
	core.C2 = make(chan *thread.Response, 10)
	core.C3 = make(chan *thread.Request, 10)
	core.C4 = make(chan *thread.Response, 10)
	core.C5 = make(chan *thread.Request, 10)
	core.C6 = make(chan *thread.Response, 10)
	core.C7 = make(chan *thread.Request, 10)
	core.C8 = make(chan *thread.Response, 10)
	core.C9 = make(chan *thread.Request, 10)
	core.C10 = make(chan *thread.Response, 10)
	core.C11 = make(chan *thread.Request, 10)
	core.C12 = make(chan *thread.Response, 10)
	core.C13 = make(chan *thread.Request, 10)
	core.C14 = make(chan *thread.Response, 10)
	core.C15 = make(chan *thread.Request, 10)
	core.C16 = make(chan *thread.Response, 10)
	core.C17 = make(chan *thread.Request, 10)
	core.C18 = make(chan *thread.Request, 10)
	core.C19 = make(chan *thread.Response, 10)
	core.C20 = make(chan *thread.Request, 10)
	core.C21 = make(chan *thread.Response, 10)
	core.C22 = make(chan *thread.Request, 10)
	core.C23 = make(chan *thread.Response, 10)
	core.C24 = make(chan *thread.Request, 10)  // free for use
	core.C25 = make(chan *thread.Response, 10) // free for use
	core.C26 = make(chan *thread.Request, 10)
	core.C27 = make(chan *thread.Response, 10)

	/* load the cfg in for the first time */
	core.config = GetConfigInstance(configPath)

	// HTTP CLIENT LOGICAL THREAD

	restLogger, err := logging.NewLogger(RestAPI.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	httpConfig := &restApi.Config{}
	core.config.FillHttpClientConfig(httpConfig)

	core.RestThread, err = restApi.New(httpConfig, restLogger,
		core.interrupt, core.C1, core.C2, core.C5, core.C6, core.C20, core.C21, core.C22, core.C23)
	if err != nil {
		return nil, err
	}

	httpProcessorLogger, err := logging.NewLogger(Socket.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	// SOCKET LOGICAL THREAD

	socketConfig := &socket.Config{}
	core.config.FillSocketConfig(socketConfig)

	core.SocketThread, err = socket.New(socketConfig, httpProcessorLogger,
		core.interrupt, core.C7, core.C8, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	// PROCESSOR LOGICAL THREAD

	processorLogger, err := logging.NewLogger(Processor.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	processorConfig := &processor.Config{}
	core.config.FillProcessorConfig(processorConfig)

	table := processorCmp.NewTable()

	core.ProcessorThread, err = processor.New(processorConfig, processorLogger, table,
		core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C11, core.C12, core.C13, core.C14, core.C18, core.C19)
	if err != nil {
		return nil, err
	}

	// SUPERVISOR LOGICAL THREAD

	runnerLogger, err := logging.NewLogger(Runner.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	runnerConfig := &runner.Config{}
	core.config.FillRunnerConfig(runnerConfig)

	registry := supervisorDb.NewLocalDatabase()

	core.RunnerThread, err = runner.NewThread(runnerConfig, runnerLogger, registry,
		core.interrupt, core.C13, core.C14, core.C15, core.C16, core.C17, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	// MESSENGER LOGICAL THREAD

	messengerLogger, err := logging.NewLogger(Messenger.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	messengerConfig := &messenger.Config{}
	core.config.FillMessengerConfig(messengerConfig)

	logger := log.New(core.config.Debug)

	core.MessengerThread, err = messenger.New(messengerConfig, messengerLogger, logger,
		core.interrupt, core.C3, core.C4, core.C17, core.C22, core.C23)
	if err != nil {
		return nil, err
	}

	// DATABASE LOGICAL THREAD

	databaseLogger, err := logging.NewLogger(Database.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	databaseConfig := &database.Config{}
	core.config.FillDatabaseConfig(databaseConfig)

	configDatabase := configDb.NewLocalPipelineDatabase()
	statDatabase := statisticDb.NewLocalStatisticDatabase()
	jobDatabase := job.NewLocalJobDatabase()

	core.DatabaseThread, err = database.New(databaseConfig, databaseLogger,
		statDatabase, configDatabase, jobDatabase,
		core.config.Paths.Configs, core.config.Paths.Statistics,
		core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C11, core.C12, core.C15, core.C16, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	// CACHE LOGICAL THREAD

	//cacheLogger, err := logging.NewLogger(Cache.ToString(), &GetConfigInstance().Debug)
	//if err != nil {
	//	return nil, err
	//}
	//
	//cacheConfig := &cache.Config{}
	//core.config.FillCacheConfig(cacheConfig)
	//
	//cacheInstance := cache_cmp.NewCache(1000)
	//
	//core.CacheThread, err = cache.New(cacheConfig, cacheLogger, cacheInstance,
	//	core.interrupt, core.C9, core.C10, core.C24, core.C25)
	//if err != nil {
	//	return nil, err
	//}

	// SCHEDULER LOGICAL THREAD

	schedulerLogger, err := logging.NewLogger(Scheduler.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	schedulerConig := &scheduler.Config{}
	core.config.FillSchedulerConfig(schedulerConig)

	core.SchedulerThread, err = scheduler.New(schedulerConig, schedulerLogger, jobDatabase,
		core.interrupt, core.C18, core.C19, core.C20, core.C21, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	// CORE DEFINITIONS

	coreLogger, err := logging.NewLogger(Undefined.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.logger = coreLogger

	return core, nil
}

const (
	Version string = "v0.22.0"
)

func (core *Core) Run() {

	core.banner()

	core.logger.SetColour(logging.Purple)

	if GetConfigInstance().Debug {
		core.logger.Println("debug mode ON")
	} else {
		core.logger.Println("debug mode OFF")
	}

	// needed in-case the proceeding thread need logging or email capabilities during startup
	core.MessengerThread.Setup()
	go core.MessengerThread.Start() // event loop
	if core.config.Debug {
		core.logger.Println("Messenger Thread Started")
	}

	// needed in-case the runner or rest thread need to populate data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if core.config.Debug {
		core.logger.Println("Database Thread Started")
	}

	// if we chain requests, we should have a way to save that data for re-use
	// FIX: the cache should start up before the provisioner in case the provisioner
	//		has stream processes that need to start using it.
	//core.CacheThread.Setup()
	//go core.CacheThread.Start()
	//if core.config.Debug {
	//	core.logger.Println("Cache Thread Started")
	//}

	core.RunnerThread.Setup()
	go core.RunnerThread.Start()
	if core.config.Debug {
		core.logger.Println("Runtime Thread Started")
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

	core.SocketThread.Setup()
	go core.SocketThread.Start()
	if core.config.Debug {
		core.logger.Println("HTTP Processor API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n",
			core.config.Net.Processor.Host, core.config.Net.Processor.Port)
	}

	// the gateway to the frontend cluster should be the last startup
	core.RestThread.Setup()
	go core.RestThread.Start() // event loop
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

	// monitor system calls being sent to the processor, if the etl is being
	// statistic on a log machine, the developer might attempt to kill the processor with SIGINT
	// requiring us to cleanly close the application without risking the loss of data
	// ---
	// an interrupt can be sent by any thread that has access to the channel if an
	// error or end-state has been reached by the application
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		core.interrupt <- thread.Panic
	case interrupt := <-core.interrupt:
		switch interrupt {
		case thread.Panic:
			core.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			core.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	core.logger.SetColour(logging.Red)

	// close the gateway, stop new thread from flooding into the servers
	core.RestThread.Teardown()

	if core.config.Debug {
		core.logger.Println("rest shutdown")
	}

	core.SocketThread.Teardown()

	if core.config.Debug {
		core.logger.Println("processor shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	//flock.ProvisionerThread.Teardown()
	//
	//if common.GetConfigInstance().Debug {
	//	flock.logger.Println("provisioner shutdown")
	//}

	core.SchedulerThread.Teardown()

	if core.config.Debug {
		core.logger.Println("scheduler shutdown")
	}

	core.ProcessorThread.Teardown()

	if core.config.Debug {
		core.logger.Println("processor shutdown")
	}

	core.RunnerThread.Teardown()

	if core.config.Debug {
		core.logger.Println("runner shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the data is useless, shutdown
	//core.CacheThread.Teardown()

	//if core.config.Debug {
	//	core.logger.Println("cache shutdown")
	//}

	// the runner might need to database data while finishing, close after
	core.DatabaseThread.Teardown()

	if core.config.Debug {
		core.logger.Println("database shutdown")
	}

	// the preceding thread might need to log, or send alerts of failure during shutdown
	core.MessengerThread.Teardown()

	if core.config.Debug {
		core.logger.Println("messenger shutdown")
	}
}
