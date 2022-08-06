package cluster

import (
	"ETLFramework/channel"
	"sync"
	"time"
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type OnCrash int8

const (
	Restart   OnCrash = 0
	DoNothing         = 1
)

type Cluster interface {
	ExtractFunc(output channel.OutputChannel)
	TransformFunc(input channel.InputChannel, output channel.OutputChannel)
	LoadFunc(input channel.InputChannel)
}

type Config struct {
	etChannelThreshold    int
	etChannelGrowthFactor int
	tlChannelThreshold    int
	tlChannelGrowthFactor int
}

type Data struct {
	numProvisionedTransformRoutes int
	numProvisionedLoadRoutines    int
}

type Monitor struct {
	group     Cluster
	config    Config
	data      Data
	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel
	waitGroup sync.WaitGroup
}

type Response struct {
	config     Config
	data       Data
	lapsedTime time.Duration
}

type Supervisor struct {
	functions map[string]Cluster
	configs   map[string]Config

	mutex sync.Mutex
}
