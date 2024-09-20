package channel

import (
	"github.com/GabeCordo/clarence/internal/interfaces"
	"sync"
	"time"
)

type Status int

const (
	Empty Status = iota
	Idle
	Healthy
	Underutilized
	Congested
)

const QueueSize = 10000

type OutputChannel chan<- any

type InputChannel <-chan any

type ManagedChannelConfig struct {
	Threshold              int
	UnderutilizedThreshold int
	GrowthFactor           float64
}

type Wrapper struct {
	In   time.Time
	Data any
}

func (w Wrapper) IsInvalid() bool {
	return w.In.IsZero() || w.Data == nil
}

type ManagedChannel struct {
	Name string

	State  Status
	Size   int
	Config ManagedChannelConfig

	Statistics     *interfaces.TimingStatistics
	TotalProcessed int

	channel chan Wrapper

	LastPush               time.Time
	UnderutilizedThreshold int
	StopNewPushes          bool
	ChannelFinished        bool

	NumOfProducers int

	producerMux sync.Mutex
	sizeMux     sync.Mutex

	wg sync.WaitGroup
}

func New(name string, threshold int, growth float64, stats *interfaces.TimingStatistics) *ManagedChannel {
	mc := new(ManagedChannel)

	mc.Name = name
	mc.Config.Threshold = threshold
	mc.Config.GrowthFactor = growth
	mc.TotalProcessed = 0

	// allocate size(Wrapper) * QueueSize in advance to accommodate
	// the incoming data to the channel
	mc.channel = make(chan Wrapper, QueueSize)
	mc.Config.UnderutilizedThreshold = mc.Config.Threshold / 3
	mc.Statistics = stats

	mc.ChannelFinished = false
	mc.StopNewPushes = false
	mc.NumOfProducers = 0

	return mc
}
