package supervisor

import (
	pipeline_cfg "github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/processor/pipeline"
	"sync"
	"time"
)

const (
	MaxConcurrentSupervisors = 24
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

//type LoadAll interface {
//	LoadFunc(helper cluster.H, metadata cluster.M, in []any)
//}
//
//type LoadOne interface {
//	LoadFunc(helper cluster.H, metadata cluster.M, in any)
//}
//
//type SystemFunctions interface {
//	Setup(curr time.Time, h cluster.H)
//	Teardown(curr time.Time, h cluster.H)
//}

type VerifiableET interface {
	VerifyETFunction(in any) (valid bool)
}

type VerifiableTL interface {
	VerifyTLFunction(in any) (valid bool)
}

// Test
// TODO : needs to be implemented
//type Test interface {
//	MockExtractFunc(metadata cluster.M, out cluster.Out)
//	VerifyTransformOutput(metadata cluster.M, in any) (success bool)
//	MockLoadFunc(metadata cluster.M, in any)
//}

type Response struct {
	Stats      *statistic.Statistics `json:"stats"`
	LapsedTime time.Duration         `json:"lapsed-time"`
	DidItCrash bool                  `json:"crashed"`
}

func NewResponse(statistics *statistic.Statistics, lapsedTime time.Duration, crashed bool) *Response {
	response := new(Response)

	response.Stats = statistics
	response.LapsedTime = lapsedTime
	response.DidItCrash = crashed

	return response
}

type Status string

const (
	UnTouched    Status = "untouched"
	Running             = "running"
	Provisioning        = "provisioning"
	Failed              = "failed"
	Stopping            = "stopping"
	Terminated          = "terminated"
	Unknown             = "-"
)

type Event uint8

const (
	Startup Event = iota
	StartProvision
	EndProvision
	Error
	Suspend
	TearedDown
	StartReport
	EndReport
)

const MaximumRoutinesPerSupervisor = 2000

type Supervisor struct {
	Id uint64 `json:"id"`

	//Pipeline *Pipeline.Pipeline    `json:"common"`

	//Stats    *statistic.Statistics `json:"stats"`

	State Status `json:"status"`
	//Mode      cluster.OnCrash       `json:"on-crash"`
	StartTime time.Time `json:"quitE-time"`

	Pipeline *pipeline.Pipeline

	metadata map[string]string

	//metadata cluster.M `json:"meta-data"`
	//helper   cluster.H

	functions []any // initialized in new
	//active    []int // ?

	//mutexes  []sync.Mutex             // ?
	//quit     []chan bool              // ?
	//channels []*duplex.ManagedChannel // initialized in new

	loadWaitGroup sync.WaitGroup
	waitGroup     sync.WaitGroup
	threadMutex   sync.Mutex
	mutex         sync.RWMutex
}

func New(config *pipeline_cfg.Pipeline, functions []any, metadata map[string]string) *Supervisor {
	supervisor := new(Supervisor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.State = UnTouched

	supervisor.functions = functions
	supervisor.Pipeline = pipeline.New(config, functions)
	supervisor.metadata = metadata

	// TODO : Pipeline creation

	// TODO : future?
	//if helper != nil {
	//	supervisor.helper = helper
	//} else {
	//	// TODO : fix
	//	panic("helper cannot be nil")
	//}
	//
	//if metadata != nil {
	//	supervisor.metadata = NewMetadata(metadata)
	//} else {
	//	supervisor.metadata = NewMetadata(nil)
	//}

	return supervisor
}

type Summary struct {
	Namespace  string
	Pipeline   string
	Supervisor uint64
	Statistics *statistic.Statistics
}
