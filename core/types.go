package core

import (
	"github.com/GabeCordo/cluster-tools/core/threads/cache"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/cluster-tools/core/threads/database"
	"github.com/GabeCordo/cluster-tools/core/threads/http_client"
	"github.com/GabeCordo/cluster-tools/core/threads/http_processor"
	"github.com/GabeCordo/cluster-tools/core/threads/messenger"
	"github.com/GabeCordo/cluster-tools/core/threads/processor"
	"github.com/GabeCordo/cluster-tools/core/threads/scheduler"
	"github.com/GabeCordo/cluster-tools/core/threads/supervisor"
	"github.com/GabeCordo/toolchain/logging"
)

type ThreadType uint8

const (
	HttpClient ThreadType = iota
	HttpProcessor
	Processor
	Supervisor
	Database
	Messenger
	Cache
	Scheduler
	Undefined
)

func (threadType ThreadType) ToString() string {
	switch threadType {
	case HttpClient:
		return "HTTP-CLIENT"
	case HttpProcessor:
		return "HTTP-PROCESSOR"
	case Processor:
		return "PROCESSOR"
	case Supervisor:
		return "SUPERVISOR"
	case Messenger:
		return "MESSENGER"
	case Database:
		return "DATABASE"
	case Cache:
		return "CACHE"
	case Scheduler:
		return "SCHEDULER"
	default:
		return "-"
	}
}

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
		core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C11, core.C12, core.C15, core.C16)
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
	core.SchedulerThread, err = scheduler.New(schedulerConig, schedulerLogger, core.interrupt, core.C18, core.C19, core.C20, core.C21)
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
