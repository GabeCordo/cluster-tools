package provisioner

import (
	"github.com/GabeCordo/ScalingFunctions"
	"sync"
	"sync/atomic"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/buffers"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/processor/component/provision"
)

const MaxNumOfSupervisors = 1

type ProvisionRequest struct {
	Namespace  string
	Supervisor uint64
	Metadata   map[string]string
	Core       string
	Pipeline   *ScalingFunctions.PipelineIR
}
type UseCases struct {
	Repository   *ScalingFunctions.Repository
	Provisioner  *provision.Provisioner
	Logger       logging.Logger
	Backlog      *buffers.RingBuffer
	activeRuns   atomic.Int64
	backlogMutex sync.RWMutex
	runWg        sync.WaitGroup // wait group on the number of active runs
}
