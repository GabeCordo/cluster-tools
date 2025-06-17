package provisioner

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/processor"
	"github.com/GabeCordo/Flock/internal/nonce"
	"github.com/GabeCordo/Flock/internal/processor/provision"
	"github.com/GabeCordo/Flock/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
)

const nonceMin = 1000000
const nonceMax = 2000000

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

	provisioner *provision.Provisioner

	requestBacklog         []threads.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	numOfActiveSupervisors int                          // tracks the number of supervisors running in the system at a time
	backlogMutex           sync.RWMutex

	noncePool *nonce.Pool

	accepting bool
	runWg     sync.WaitGroup // wait group on the number of active runs
	requestWg sync.WaitGroup // wait group on the number of processed async messages
}

func NewThread(cfg *Config, logger *logging.Logger, provisioner *provision.Provisioner, channels ...interface{}) (*Thread, error) {
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

	instance.provisioner = provisioner

	instance.requestBacklog = make([]threads.ProvisionerRequest, 0)
	instance.numOfActiveSupervisors = 0

	instance.noncePool = nonce.New(nonceMin, nonceMax)

	instance.logger.SetColour(logging.Orange)

	return instance, nil
}
