package channel

import (
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

type DataTimer struct {
	In  time.Time
	Out time.Time
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
	Timestamps     map[uint64]DataTimer

	channel chan DataWrapper

	LastPush        time.Time
	ChannelFinished bool

	mutex sync.Mutex
	wg    sync.WaitGroup
}

type OneWayManagedChannel struct {
	channel *ManagedChannel
}
