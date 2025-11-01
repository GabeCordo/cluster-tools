package core

import (
	"github.com/FortifiedCode/flock/internal/shared/drivers/mongo"
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/shared/nonce"
	"github.com/FortifiedCode/flock/internal/shared/socket/json_socket"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
	"github.com/FortifiedCode/flock/internal/targets/core/component/message/log"
	processorCmp "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	"github.com/FortifiedCode/flock/internal/targets/core/component/scheduler/job"
	jobDb "github.com/FortifiedCode/flock/internal/targets/core/database/job/mongo"
	pipelineDb "github.com/FortifiedCode/flock/internal/targets/core/database/pipeline/mongo"
	runDb "github.com/FortifiedCode/flock/internal/targets/core/database/run/in_memory"
	statisticDb "github.com/FortifiedCode/flock/internal/targets/core/database/statistic/mongo"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/database"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/messenger"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/processor"
	restApi "github.com/FortifiedCode/flock/internal/targets/core/thread/rest"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/runner"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/scheduler"
	"github.com/FortifiedCode/flock/internal/targets/core/thread/socket"
	databaseUc "github.com/FortifiedCode/flock/internal/targets/core/use_cases/database"
	processorUc "github.com/FortifiedCode/flock/internal/targets/core/use_cases/processor"
	runnerUc "github.com/FortifiedCode/flock/internal/targets/core/use_cases/runner"
	schedulerUc "github.com/FortifiedCode/flock/internal/targets/core/use_cases/scheduler"
	socketUc "github.com/FortifiedCode/flock/internal/targets/core/use_cases/socket"
	"os"
	"os/signal"
	"syscall"

	"github.com/FortifiedCode/flock/internal/shared/logging/text_logging"
)

const (
	restNonceMin      = 0
	restNonceMax      = 262114
	schedulerNonceMin = 262114
	schedulerNonceMax = 524228 // (base) 262114 + 262114 (offset)
	socketNonceMin    = 524228
	socketNonceMax    = 786342 // (base) 524228 + 262114 (offset)
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
	logger logging.Logger
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

	restLogger, err := text_logging.New(RestAPI.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	httpConfig := &restApi.Config{}
	core.config.FillHttpClientConfig(httpConfig)

	restNoncePool := nonce.New(restNonceMin, restNonceMax)

	core.RestThread, err = restApi.New(httpConfig, restLogger, restNoncePool,
		core.interrupt, core.C1, core.C2, core.C5, core.C6, core.C20, core.C21, core.C22, core.C23)
	if err != nil {
		return nil, err
	}

	// SOCKET LOGICAL THREAD

	socketConfig := &socket.Config{}
	core.config.FillSocketConfig(socketConfig)

	socketLogger, err := text_logging.New(Socket.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	socketNoncePool := nonce.New(socketNonceMin, socketNonceMax)

	jsonSocket := json_socket.NewServer()

	socketUseCases := socketUc.UseCases{Socket: jsonSocket, Logger: socketLogger}

	core.SocketThread, err = socket.New(socketConfig, socketLogger, socketNoncePool, &socketUseCases,
		core.interrupt, core.C7, core.C8, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	// PROCESSOR LOGICAL THREAD

	processorLogger, err := text_logging.New(Processor.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	processorConfig := &processor.Config{}
	core.config.FillProcessorConfig(processorConfig)

	table := processorCmp.NewTable()

	processorUseCases := processorUc.UseCases{
		ProcessorTable: table,
		Logger:         processorLogger,
	}

	core.ProcessorThread, err = processor.New(processorConfig, processorLogger, processorUseCases,
		core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C11, core.C12, core.C13, core.C14, core.C18, core.C19)
	if err != nil {
		return nil, err
	}

	// SUPERVISOR LOGICAL THREAD

	runnerLogger, err := text_logging.New(Runner.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	runnerConfig := &runner.Config{}
	core.config.FillRunnerConfig(runnerConfig)

	registry := runDb.NewLocalDatabase()

	runnerUseCases := runnerUc.UseCases{
		RunDatabase: registry,
	}

	core.RunnerThread, err = runner.NewThread(runnerConfig, runnerLogger, runnerUseCases,
		core.interrupt, core.C13, core.C14, core.C15, core.C16, core.C17, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	// MESSENGER LOGICAL THREAD

	messengerLogger, err := text_logging.New(Messenger.ToString(), &GetConfigInstance().Debug)
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

	databaseLogger, err := text_logging.New(Database.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	databaseConfig := &database.Config{}
	core.config.FillDatabaseConfig(databaseConfig)

	envVars := ReadEnvironmentVariables()
	driver := mongo.NewDriver(envVars.MongoDbUri)
	err = driver.Connect()
	if err != nil {
		return nil, err
	}

	configDatabase, err := pipelineDb.NewMongoDatabase(driver)
	if err != nil {
		return nil, err
	}

	statDatabase, err := statisticDb.NewMongoDatabase(driver)
	if err != nil {
		return nil, err
	}

	jobDatabase, err := jobDb.NewMongoDatabase(driver)
	if err != nil {
		return nil, err
	}

	databaseUseCases := databaseUc.UseCases{
		PipelineDatabase:  configDatabase,
		StatisticDatabase: statDatabase,
		JobDatabase:       jobDatabase,
		Logger:            databaseLogger,
	}

	core.DatabaseThread, err = database.New(databaseConfig, databaseLogger, databaseUseCases,
		core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C11, core.C12, core.C15, core.C16, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	// CACHE LOGICAL THREAD

	//cacheLogger, err := logging.NewLogger(CacheComponent.ToString(), &GetConfigInstance().Debug)
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
	//	core.interrupt, core.c9, core.c10, core.c24, core.c25)
	//if err != nil {
	//	return nil, err
	//}

	// SCHEDULER LOGICAL THREAD

	schedulerLogger, err := text_logging.New(Scheduler.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}

	schedulerConfig := &scheduler.Config{}
	core.config.FillSchedulerConfig(schedulerConfig)

	sch, err := job.New(jobDatabase)
	if err != nil {
		return nil, err
	}

	schedulerUseCases := schedulerUc.UseCases{
		Scheduler:    sch,
		JobsDatabase: sch.Jobs,
		Logger:       schedulerLogger,
	}

	scheduleNoncePool := nonce.New(schedulerNonceMin, schedulerNonceMax)

	core.SchedulerThread, err = scheduler.New(schedulerConfig, schedulerLogger, schedulerUseCases, scheduleNoncePool,
		core.interrupt, core.C18, core.C19, core.C20, core.C21, core.C26, core.C27)
	if err != nil {
		return nil, err
	}

	// CORE DEFINITIONS

	coreLogger, err := text_logging.New(Undefined.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.logger = coreLogger

	return core, nil
}

func (core *Core) Run() {

	core.logger.SetColour(terminal.Purple)

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
	//	core.logger.Println("CacheComponent Thread Started")
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
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sig:
		core.interrupt <- thread.Shutdown
	case i := <-core.interrupt:
		switch i {
		case thread.Panic:
			core.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			core.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	core.logger.SetColour(terminal.Red)

	// close the gateway, stop new thread from flooding into the servers
	core.RestThread.Teardown()

	if core.config.Debug {
		core.logger.Println("rest shutdown")
	}

	core.SocketThread.TearDown()

	if core.config.Debug {
		core.logger.Println("processor shutdown")
	}

	core.SchedulerThread.TearDown()

	if core.config.Debug {
		core.logger.Println("scheduler shutdown")
	}

	core.ProcessorThread.Teardown()

	if core.config.Debug {
		core.logger.Println("processor shutdown")
	}

	core.RunnerThread.TearDown()

	if core.config.Debug {
		core.logger.Println("runner shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the data is useless, shutdown
	//core.CacheThread.TearDown()

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
