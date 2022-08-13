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
	Identifier            string `json:"identifier"`
	etChannelThreshold    int    `json:"et-channel-threshold"`
	etChannelGrowthFactor int    `json:"et-channel-growth-factor"`
	tlChannelThreshold    int    `json:"tl-channel-threshold"`
	tlChannelGrowthFactor int    `json:"tl-channel-growth-factor"`
}

type Statistics struct {
	NumProvisionedTransformRoutes int `json:"num-provisioned-transform-routes"`
	NumProvisionedLoadRoutines    int `json:"num-provisioned-load-routines"`
}

type Monitor struct {
	group     Cluster     `json:"group"`
	Config    *Config     `json:"config"`
	Stats     *Statistics `json:"stats"`
	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel
	waitGroup sync.WaitGroup
}

type Response struct {
	Config     *Config       `json:"config"`
	Stats      *Statistics   `json:"stats""`
	LapsedTime time.Duration `json:"lapsed-time"`
}

type Supervisor struct {
	Functions map[string]Cluster `json:"functions"`
	Configs   map[string]Config  `json:"configs"`

	mutex sync.Mutex
}
