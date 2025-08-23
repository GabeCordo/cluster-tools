package duplex

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/FortifiedCode/flock/internal/core/database/statistic"
)

type BadManagedChannelType struct {
	description string
}

func (bmce BadManagedChannelType) Error() string {
	return bmce.description
}

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
	Data []reflect.Value
}

func (w Wrapper) IsInvalid() bool {
	return w.In.IsZero() || w.Data == nil
}

type ManagedChannel struct {
	Name string

	State  Status
	Size   int
	Config ManagedChannelConfig

	Statistics     *statistic.TimingStatistics
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

func New(name string, threshold int, growth float64, stats *statistic.TimingStatistics) *ManagedChannel {
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

func (status Status) ToString() string {

	switch status {
	case Idle:
		return "Idle"
	case Empty:
		return "Empty"
	case Congested:
		return "Congested"
	default:
		return "Healthy"
	}
}

func (mc *ManagedChannel) GetChannel() chan Wrapper {
	return mc.channel
}

func (mc *ManagedChannel) Push(data []reflect.Value) bool {

	mc.sizeMux.Lock()

	// don't push to the channel if it is supposed to be closed
	if mc.StopNewPushes {
		return false
	}

	// see if we are hitting a threshold and the successive function is
	// getting overloaded with data units
	if (mc.Size + 1) >= mc.Config.Threshold {
		mc.State = Congested
	}

	mc.Size++
	mc.TotalProcessed++

	currentTime := time.Now()
	mc.LastPush = currentTime

	// before writing on the channel we want to release the mutex otherwise
	// the managed channel can enter a deadlock;
	//
	// producer: wants to send more data onto the ManagedChannel but has to
	//			 wait until the channel has buffer to allow it
	//			 -> channel send is blocked
	//
	// consumer: wants to read data from the ManagedChannel and decrement the
	//			 size but it needs to enter the mutex to do that.
	//			 -> the mutex on DataPopped() will block waiting for the chance
	//				to write to the queue
	//				BUT
	//				since the producer is waiting to write to the channel and
	//				the consumer can't continue pulling data off the queue we
	//				enter an infinite deadlock!
	mc.sizeMux.Unlock()

	if data == nil {
		return false
	}

	mc.channel <- Wrapper{In: currentTime, Data: data}

	return true
}

func (mc *ManagedChannel) DataPopped(timeIntoQueue time.Time) {

	mc.sizeMux.Lock()
	defer mc.sizeMux.Unlock()

	timeOutOfQueue := time.Now()
	totalTimeInQueue := timeOutOfQueue.Sub(timeIntoQueue)

	if mc.Statistics.AverageTime != 0 {
		if totalTimeInQueue > mc.Statistics.MaxTimeBeforePop {
			mc.Statistics.MaxTimeBeforePop = totalTimeInQueue
		} else if totalTimeInQueue < mc.Statistics.MinTimeBeforePop {
			mc.Statistics.MinTimeBeforePop = totalTimeInQueue
		}
		mc.Statistics.AverageTime += totalTimeInQueue / 2
	} else {
		mc.Statistics.AverageTime = totalTimeInQueue
		mc.Statistics.MedianTime = 0 // TODO: support
		mc.Statistics.MaxTimeBeforePop = totalTimeInQueue
		mc.Statistics.MinTimeBeforePop = totalTimeInQueue
	}

	mc.Size--
	mc.State = mc.GetState()
}

func (mc *ManagedChannel) Accepting() bool {
	return !mc.StopNewPushes
}

func (mc *ManagedChannel) StopPushes() {
	mc.StopNewPushes = true
}

func (mc *ManagedChannel) AddProducer() {
	mc.producerMux.Lock()
	defer mc.producerMux.Unlock()

	mc.NumOfProducers++
}

func (mc *ManagedChannel) ProducerDone() {

	mc.producerMux.Lock()
	defer mc.producerMux.Unlock()

	mc.NumOfProducers--

	if !mc.ChannelFinished && (mc.NumOfProducers <= 0) {
		mc.ChannelFinished = true
		close(mc.channel)
	}
}

func (mc *ManagedChannel) GetState() Status {

	if mc.Size == 0 {
		if time.Now().Sub(mc.LastPush).Seconds() > 3 {
			mc.State = Idle
		} else {
			mc.State = Empty
		}
	} else if ((mc.State == Congested) || (mc.State == Idle)) && (mc.Size < mc.Config.UnderutilizedThreshold) {
		mc.State = Underutilized
	} else if mc.Size > mc.Config.Threshold {
		mc.State = Congested
	} else {
		mc.State = Healthy
	}

	return mc.State
}

func (mc *ManagedChannel) GetGrowthFactor() float64 {

	return mc.Config.GrowthFactor
}

func (mc *ManagedChannel) AmountOfDataSeen() int {
	return mc.TotalProcessed
}

func (mc *ManagedChannel) ToString() string {
	return fmt.Sprintf("[%s][%s][Size: %d]\n", mc.Name, mc.State.ToString(), mc.Size)
}
