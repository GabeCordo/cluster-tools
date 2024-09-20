package provisioner

import (
	"errors"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"github.com/GabeCordo/clarence/internal/threads/common"
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

	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan common.ProvisionerRequest    // Supervisor is receiving threads from the http_thread
	C2 chan<- common.ProvisionerResponse // Supervisor is sending responses to the http_thread

	logger *logging.Logger

	requestBacklog         []common.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	numOfActiveSupervisors int                         // tracks the number of supervisors running in the system at a time
	backlogMutex           sync.RWMutex

	accepting   bool
	listenersWg sync.WaitGroup
	requestWg   sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, channels ...interface{}) (*Thread, error) {
	provisioner := new(Thread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	provisioner.C1, ok = (channels[1]).(chan common.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	provisioner.C2, ok = (channels[2]).(chan common.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	provisioner.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.Config type")
	}
	provisioner.Config = cfg

	provisioner.requestBacklog = make([]common.ProvisionerRequest, 0)
	provisioner.numOfActiveSupervisors = 0

	provisioner.logger.SetColour(logging.Orange)

	return provisioner, nil
}
