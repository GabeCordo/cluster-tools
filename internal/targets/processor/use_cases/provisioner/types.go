package provisioner

import (
	"github.com/FortifiedCode/flock/internal/shared/buffers"
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/targets/processor/component/provision"
	"github.com/FortifiedCode/plover"
	"sync"
	"sync/atomic"
)

const MaxNumOfSupervisors = 1

type ProvisionRequest struct {
	Namespace  string
	Supervisor uint64
	Metadata   map[string]string
	Core       string
	Pipeline   *plover.PipelineIR
}
type UseCases struct {
	Repository   *plover.Repository
	Provisioner  *provision.Provisioner
	Logger       logging.Logger
	Backlog      *buffers.RingBuffer
	activeRuns   atomic.Int64
	backlogMutex sync.RWMutex
	runWg        sync.WaitGroup // wait group on the number of active runs
}
