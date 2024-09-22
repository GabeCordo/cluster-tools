package supervisor

import (
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel/duplex"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
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

type LoadAll interface {
	LoadFunc(helper cluster.H, metadata cluster.M, in []any)
}

type LoadOne interface {
	LoadFunc(helper cluster.H, metadata cluster.M, in any)
}

type SystemFunctions interface {
	Setup(curr time.Time, h cluster.H)
	Teardown(curr time.Time, h cluster.H)
}

type VerifiableET interface {
	VerifyETFunction(in any) (valid bool)
}

type VerifiableTL interface {
	VerifyTLFunction(in any) (valid bool)
}

// Test
// TODO : needs to be implemented
type Test interface {
	MockExtractFunc(metadata cluster.M, out cluster.Out)
	VerifyTransformOutput(metadata cluster.M, in any) (success bool)
	MockLoadFunc(metadata cluster.M, in any)
}

type Response struct {
	Config     cluster.Config         `json:"core"`
	Stats      *interfaces.Statistics `json:"stats"`
	LapsedTime time.Duration          `json:"lapsed-time"`
	DidItCrash bool                   `json:"crashed"`
}

func NewResponse(config cluster.Config, statistics *interfaces.Statistics, lapsedTime time.Duration, crashed bool) *Response {
	response := new(Response)

	response.Config = config
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

	Config    cluster.Config         `json:"common"`
	Stats     *interfaces.Statistics `json:"stats"`
	State     Status                 `json:"status"`
	Mode      cluster.OnCrash        `json:"on-crash"`
	StartTime time.Time              `json:"quitE-time"`

	Metadata cluster.M `json:"meta-data"`

	group     cluster.Cluster
	helper    cluster.H
	ETChannel *duplex.ManagedChannel
	mutexET   sync.Mutex
	TLChannel *duplex.ManagedChannel
	mutexTL   sync.Mutex

	ActiveTRoutines int
	quitT           []chan bool

	ActiveLRoutines int
	quitL           []chan bool

	loadWaitGroup sync.WaitGroup
	waitGroup     sync.WaitGroup
	threadMutex   sync.Mutex
	mutex         sync.RWMutex
}

func NewSupervisor(clusterImplementation cluster.Cluster, metadata map[string]string, helper cluster.H) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.group = clusterImplementation
	supervisor.State = UnTouched
	supervisor.Config = cluster.DefaultConfig

	et := interfaces.NewTimingStatistics()
	tl := interfaces.NewTimingStatistics()
	supervisor.Stats = interfaces.NewStatistics(et, tl)

	supervisor.ETChannel = duplex.New("ETChannel", supervisor.Config.ETChannelThreshold, supervisor.Config.ETChannelGrowthFactor, et)
	supervisor.TLChannel = duplex.New("TLChannel", supervisor.Config.TLChannelThreshold, supervisor.Config.TLChannelGrowthFactor, tl)

	supervisor.ActiveTRoutines = 0
	supervisor.quitT = make([]chan bool, 0)
	supervisor.quitL = make([]chan bool, 0)

	if helper != nil {
		supervisor.helper = helper
	} else {
		// TODO : I could not be arsed, clean it up later, this stinks
		panic("helper can not be nil")
	}

	if metadata != nil {
		supervisor.Metadata = NewMetadata(metadata)
	} else {
		supervisor.Metadata = NewMetadata(nil)
	}
	return supervisor
}

func NewCustomSupervisor(clusterImplementation cluster.Cluster, config *cluster.Config, metadata map[string]string, helper cluster.H) *Supervisor {
	supervisor := new(Supervisor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.State = UnTouched
	supervisor.group = clusterImplementation
	supervisor.Config = *config // copy pipeline

	et := interfaces.NewTimingStatistics()
	tl := interfaces.NewTimingStatistics()
	supervisor.Stats = interfaces.NewStatistics(et, tl)

	supervisor.ETChannel = duplex.New("ETChannel", config.ETChannelThreshold, config.ETChannelGrowthFactor, et)
	supervisor.TLChannel = duplex.New("TLChannel", config.TLChannelThreshold, config.TLChannelGrowthFactor, tl)

	supervisor.ActiveTRoutines = 0
	supervisor.quitT = make([]chan bool, 0)
	supervisor.quitL = make([]chan bool, 0)

	if helper != nil {
		supervisor.helper = helper
	} else {
		// TODO : fix
		panic("helper cannot be nil")
	}

	if metadata != nil {
		supervisor.Metadata = NewMetadata(metadata)
	} else {
		supervisor.Metadata = NewMetadata(nil)
	}

	return supervisor
}

type Summary struct {
	Module     string
	Cluster    string
	Supervisor uint64
	Statistics *statistic.Statistics
	ETState    string
	ETSize     int
	TLState    string
	TLSize     int
}
