package cluster

import (
	"ETLFramework/channel"
	"sync"
	"time"
)

const (
	MaxConcurrentMonitors = 24
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type Cluster interface {
	ExtractFunc(output channel.OutputChannel)
	TransformFunc(input channel.InputChannel, output channel.OutputChannel)
	LoadFunc(input channel.InputChannel)
}

type Config struct {
	Identifier            string  `json:"identifier"`
	Mode                  OnCrash `json:"on-crash"`
	ETChannelThreshold    int     `json:"et-channel-threshold"`
	ETChannelGrowthFactor int     `json:"et-channel-growth-factor"`
	TLChannelThreshold    int     `json:"tl-channel-threshold"`
	TLChannelGrowthFactor int     `json:"tl-channel-growth-factor"`
}

type Statistics struct {
	NumProvisionedExtractRoutines int `json:"num-provisioned-extract-routines"`
	NumProvisionedTransformRoutes int `json:"num-provisioned-transform-routes"`
	NumProvisionedLoadRoutines    int `json:"num-provisioned-load-routines"`
	NumEtThresholdBreaches        int `json:"num-et-threshold-breaches"`
	NumTlThresholdBreaches        int `json:"num-tl-threshold-breaches"`
}

type OnCrash int8

const (
	Restart   OnCrash = 0
	DoNothing         = 1
)

type Status uint8

const (
	UnTouched  Status = 0
	Running           = 1
	Failed            = 2
	Terminated        = 3
)

type Event uint8

const (
	Startup        Event = 0
	Provisioning         = 1
	TeardownCalled       = 2
	TearedDown           = 3
)

type Supervisor struct {
	Id        uint32      `json:"id"`
	group     Cluster     `json:"group"`
	Config    *Config     `json:"config"`
	Stats     *Statistics `json:"stats"`
	State     Status      `json:"status"`
	mode      OnCrash     `json:"on-crash"`
	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel
	waitGroup sync.WaitGroup
}

type Response struct {
	Config     *Config       `json:"config"`
	Stats      *Statistics   `json:"stats""`
	LapsedTime time.Duration `json:"lapsed-time"`
}

type Provisioner struct {
	Functions map[string]Cluster `json:"functions"`
	Configs   map[string]Config  `json:"configs"`

	mutex sync.Mutex
}
