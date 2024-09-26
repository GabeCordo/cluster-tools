package provisioner

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/processor/provisioner"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
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

	C1 chan threads.ProvisionerRequest    // Run is receiving threads from the http_thread
	C2 chan<- threads.ProvisionerResponse // Run is sending responses to the http_thread

	logger *logging.Logger

	provisioner *provisioner.Provisioner

	requestBacklog         []threads.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	numOfActiveSupervisors int                          // tracks the number of supervisors running in the system at a time
	backlogMutex           sync.RWMutex

	accepting   bool
	listenersWg sync.WaitGroup
	requestWg   sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, provisioner *provisioner.Provisioner, channels ...interface{}) (*Thread, error) {
	instance := new(Thread)
	var ok bool

	instance.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	instance.C1, ok = (channels[1]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	instance.C2, ok = (channels[2]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	instance.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.pipeline type")
	}
	instance.Config = cfg

	instance.provisioner = provisioner

	instance.requestBacklog = make([]threads.ProvisionerRequest, 0)
	instance.numOfActiveSupervisors = 0

	instance.logger.SetColour(logging.Orange)

	return instance, nil
}
