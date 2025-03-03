package provisioner

import (
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/Sentmint/pops/internal/core/processor"
	"github.com/Sentmint/pops/internal/processor/threads"
	"github.com/Sentmint/yule"
	"sync"
)

const MaxNumOfSupervisors = 1

type Config struct {
	Debug      *bool
	Timeout    *float64
	Standalone *bool
	Core       *string
	Processor  processor.Config
}

type Thread struct {
	Config *Config

	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C0 chan threads.SocketRequest
	C1 chan threads.ProvisionerRequest    // Runtime is receiving threads from the http_thread
	C2 chan<- threads.ProvisionerResponse // Runtime is sending responses to the http_thread

	logger *logging.Logger

	repository        *yule.Repository
	repositoryPresent bool

	runnable        yule.RunnablePipeline
	injectables     []any
	runnablePresent bool

	requestBacklog []threads.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	backlogMutex   sync.RWMutex

	runnablePipelines  []yule.RunnablePipeline
	numOfActiveRunners int // tracks the number of runners active on the system at a time

	accepting bool
	runWg     sync.WaitGroup // wait group on the number of active runs
	requestWg sync.WaitGroup // wait group on the number of processed async messages
}

func NewThread(cfg *Config, logger *logging.Logger, repository *yule.Repository, runnable *yule.RunnablePipeline, injectables []any, channels ...interface{}) (*Thread, error) {
	instance := new(Thread)
	var ok bool

	instance.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	instance.C0, ok = (channels[1]).(chan threads.SocketRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SocketRequest' in index 1")
	}
	instance.C1, ok = (channels[2]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 2")
	}
	instance.C2, ok = (channels[3]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 3")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	instance.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.pipeline type")
	}
	instance.Config = cfg

	if repository != nil {
		instance.repository = repository
		instance.repositoryPresent = true
	}

	if runnable != nil {
		instance.runnable = *runnable
		instance.injectables = injectables
		instance.runnablePresent = true
	}

	instance.requestBacklog = make([]threads.ProvisionerRequest, 0)

	instance.runnablePipelines = make([]yule.RunnablePipeline, 0)
	instance.numOfActiveRunners = 0

	instance.logger.SetColour(logging.Orange)

	return instance, nil
}
