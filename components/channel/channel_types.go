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

type ManagedChannel struct {
	Name string

	State  Status
	Size   int
	Config ManagedChannelConfig

	channel chan any

	LastPush        time.Time
	ChannelFinished bool

	mutex sync.Mutex
	wg    sync.WaitGroup
}

type OneWayManagedChannel struct {
	channel *ManagedChannel
}
