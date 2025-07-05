package provisioner

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/processor/component/provision"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
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

	channels struct {
		Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		C0 chan *thread.SocketRequest
		C1 chan *thread.ProvisionerRequest    // Runtime is receiving thread from the http_thread
		C2 chan<- *thread.ProvisionerResponse // Runtime is sending responses to the http_thread

		close chan thread.InterruptEvent
	}

	logger *logging.Logger

	provisioner *provision.Provisioner

	requestBacklog         []*thread.ProvisionerRequest // a backlog of provision requests we want to avoid congesting the server
	numOfActiveSupervisors int                          // tracks the number of supervisors running in the system at a time
	backlogMutex           sync.RWMutex

	noncePool *nonce.Pool

	runWg     sync.WaitGroup // wait group on the number of active runs
	requestWg sync.WaitGroup // wait group on the number of processed async messages
}

func NewThread(cfg *Config, logger *logging.Logger, provisioner *provision.Provisioner, channels ...interface{}) (*Thread, error) {
	t := new(Thread)
	var ok bool

	t.channels.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.channels.C0, ok = (channels[1]).(chan *thread.SocketRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SocketRequest' in index 1")
	}
	t.channels.C1, ok = (channels[2]).(chan *thread.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 2")
	}
	t.channels.C2, ok = (channels[3]).(chan *thread.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 3")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.pipeline type")
	}
	t.Config = cfg

	t.provisioner = provisioner

	t.requestBacklog = make([]*thread.ProvisionerRequest, 0)
	t.numOfActiveSupervisors = 0

	t.noncePool = nonce.New(nonceMin, nonceMax)

	t.logger.SetColour(logging.Orange)

	return t, nil
}
