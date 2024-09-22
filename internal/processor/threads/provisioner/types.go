package provisioner

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
	"sync"
)

const MaxNumOfSupervisors = 5

type Config struct {
	Debug      bool
	Timeout    float64
	Standalone bool
	Core       string
	Processor  interfaces.ProcessorConfig
}

type Thread struct {
	Config *Config

	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan threads.ProvisionerRequest    // Supervisor is receiving threads from the http_thread
	C2 chan<- threads.ProvisionerResponse // Supervisor is sending responses to the http_thread

	logger *logging.Logger

	requestBacklog         []threads.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	numOfActiveSupervisors int                          // tracks the number of supervisors running in the system at a time
	backlogMutex           sync.RWMutex

	accepting   bool
	listenersWg sync.WaitGroup
	requestWg   sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, channels ...interface{}) (*Thread, error) {
	provisioner := new(Thread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	provisioner.C1, ok = (channels[1]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	provisioner.C2, ok = (channels[2]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	provisioner.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.pipeline type")
	}
	provisioner.Config = cfg

	provisioner.requestBacklog = make([]threads.ProvisionerRequest, 0)
	provisioner.numOfActiveSupervisors = 0

	provisioner.logger.SetColour(logging.Orange)

	return provisioner, nil
}
