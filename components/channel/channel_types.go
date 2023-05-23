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

type Message any

type OutputChannel chan<- Message

type InputChannel <-chan Message

type ManagedChannelConfig struct {
	Threshold    int
	GrowthFactor int
	Size         int
}

type ManagedChannel struct {
	name string

	state    Status
	config   ManagedChannelConfig
	Channel  chan Message
	lastPush time.Time

	channelFinished bool

	mutex sync.Mutex
	wg    sync.WaitGroup
}
