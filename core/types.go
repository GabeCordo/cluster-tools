package core

import (
	"github.com/GabeCordo/mango/core/threads/cache"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/mango/core/threads/database"
	"github.com/GabeCordo/mango/core/threads/http_client"
	"github.com/GabeCordo/mango/core/threads/http_processor"
	"github.com/GabeCordo/mango/core/threads/messenger"
	"github.com/GabeCordo/mango/core/threads/processor"
	"github.com/GabeCordo/mango/core/threads/scheduler"
	"github.com/GabeCordo/mango/core/threads/supervisor"
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

	C1        chan common.DatabaseRequest
	C2        chan common.DatabaseResponse
	C3        chan common.MessengerRequest
	C4        chan common.MessengerResponse
	C5        chan common.ProcessorRequest
	C6        chan common.ProcessorResponse
	C7        chan common.ProcessorRequest
	C8        chan common.ProcessorResponse
	C9        chan common.CacheRequest
	C10       chan common.CacheResponse
	C11       chan common.DatabaseRequest
	C12       chan common.DatabaseResponse
	C13       chan common.SupervisorRequest
	C14       chan common.SupervisorResponse
	C15       chan common.DatabaseRequest
	C16       chan common.DatabaseResponse
	C17       chan common.MessengerRequest
	C18       chan common.ProcessorRequest
	C19       chan common.ProcessorResponse
	interrupt chan common.InterruptEvent

	config *Config
	logger *logging.Logger
}

func New(configPath string) (*Core, error) {
	core := new(Core)

	core.interrupt = make(chan common.InterruptEvent, 10)
	core.C1 = make(chan common.DatabaseRequest, 10)
	core.C2 = make(chan common.DatabaseResponse, 10)
	core.C3 = make(chan common.MessengerRequest, 10)
	core.C4 = make(chan common.MessengerResponse, 10)
	core.C5 = make(chan common.ProcessorRequest, 10)
	core.C6 = make(chan common.ProcessorResponse, 10)
	core.C7 = make(chan common.ProcessorRequest, 10)
	core.C8 = make(chan common.ProcessorResponse, 10)
	core.C9 = make(chan common.CacheRequest, 10)
	core.C10 = make(chan common.CacheResponse, 10)
	core.C11 = make(chan common.DatabaseRequest, 10)
	core.C12 = make(chan common.DatabaseResponse, 10)
	core.C13 = make(chan common.SupervisorRequest, 10)
	core.C14 = make(chan common.SupervisorResponse, 10)
	core.C15 = make(chan common.DatabaseRequest, 10)
	core.C16 = make(chan common.DatabaseResponse, 10)
	core.C17 = make(chan common.MessengerRequest, 10)
	core.C18 = make(chan common.ProcessorRequest, 10)
	core.C19 = make(chan common.ProcessorResponse, 10)

	/* load the cfg in for the first time */
	core.config = GetConfigInstance(configPath)

	httpLogger, err := logging.NewLogger(HttpClient.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	httpConfig := &http_client.Config{}
	core.config.FillHttpClientConfig(httpConfig)
	core.HttpClientThread, err = http_client.New(httpConfig, httpLogger,
		core.interrupt, core.C1, core.C2, core.C5, core.C6)
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
		core.interrupt, core.C3, core.C4, core.C17)
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
		core.interrupt, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	schedulerLogger, err := logging.NewLogger(Scheduler.ToString(), &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	schedulerConig := &scheduler.Config{}
	core.config.FillSchedulerConfig(schedulerConig)
	core.SchedulerThread, err = scheduler.New(schedulerConig, schedulerLogger, core.interrupt, core.C18, core.C19)
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
