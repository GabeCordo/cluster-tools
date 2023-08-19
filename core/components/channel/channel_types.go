package channel

import (
	channel "github.com/GabeCordo/mango/components/channel"
	"sync"
	"time"
)

type Status int

const (
	Empty Status = iota
	Idle
	Healthy
	Congested
)

type OutputChannel chan<- any

type InputChannel <-chan any

type ManagedChannelConfig struct {
	Threshold    int
	GrowthFactor int
}

type DataWrapper struct {
	Id   uint64
	Data any
}

type ManagedChannel struct {
	Name string

	State  Status
	Size   int
	Config ManagedChannelConfig

	TotalProcessed int
	Timestamps     map[uint64]channel.DataTimer

	channel chan DataWrapper

	LastPush        time.Time
	StopNewPushes   bool
	ChannelFinished bool

	mutex sync.Mutex
	wg    sync.WaitGroup
}

type OneWayManagedChannel struct {
	channel *ManagedChannel
}
