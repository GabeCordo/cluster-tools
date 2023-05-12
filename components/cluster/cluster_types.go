package cluster

import (
	"github.com/GabeCordo/etl/components/channel"
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
	UnTouched    Status = 0
	Running             = 1
	Provisioning        = 2
	Failed              = 4
	Terminated          = 5
	Unknown             = 6
)

type Event uint8

const (
	Startup        Event = 0
	StartProvision       = 1
	EndProvision         = 2
	Error                = 3
	TearedDown           = 4
	StartReport          = 5
	EndReport            = 6
)

type SupervisorData struct {
	Id        uint64    `json:"id"`
	State     Status    `json:"state"`
	Mode      OnCrash   `json:"on-crash"`
	StartTime time.Time `json:"start-time"`
}

type Supervisor struct {
	Id        uint64      `json:"id"`
	group     Cluster     `json:"group"`
	Config    *Config     `json:"config"`
	Stats     *Statistics `json:"stats"`
	State     Status      `json:"status"`
	mode      OnCrash     `json:"on-crash"`
	StartTime time.Time   `json:"start-time"`

	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex
}

type SupervisorV2 struct {
	Data   SupervisorData `json:"data"`
	Config *Config        `json:"config"`
	Stats  *Statistics    `json:"stats"`

	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex
}

type Response struct {
	Config     *Config       `json:"config"`
	Stats      *Statistics   `json:"stats""`
	LapsedTime time.Duration `json:"lapsed-time"`
}

type IdentifierRegistryPair struct {
	Identifier string
	Registry   *Registry
}

type Registry struct {
	Supervisors map[uint64]*Supervisor

	idReference uint64
	mutex       sync.RWMutex
}

type Provisioner struct {
	RegisteredFunctions  map[string]*Cluster `json:"functions"`
	OperationalFunctions map[string]*Cluster
	Configs              map[string]Config    `json:"configs"`
	Registries           map[string]*Registry `json:"registries"`

	mutex sync.RWMutex
}
