package provisioner

import (
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/processor/component/provision"
	"github.com/GabeCordo/Flock/internal/shared/buffers"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"sync"
	"sync/atomic"
)

const MaxNumOfSupervisors = 1

type ProvisionRequest struct {
	Namespace  string
	Supervisor uint64
	Metadata   map[string]string
	Core       string
	Pipeline   *pipeline.Pipeline
}
type UseCases struct {
	Provisioner  *provision.Provisioner
	Logger       logging.Logger
	Backlog      *buffers.RingBuffer
	activeRuns   atomic.Int64
	backlogMutex sync.RWMutex
	runWg        sync.WaitGroup // wait group on the number of active runs
}
