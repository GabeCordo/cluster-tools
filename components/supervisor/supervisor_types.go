package supervisor

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl/components/channel"
	"sync"
	"time"
)

const (
	MaxConcurrentSupervisors = 24
)

type Status uint8

const (
	UnTouched Status = iota
	Running
	Provisioning
	Failed
	Terminated
	Unknown
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

type Supervisor struct {
	Id uint64 `json:"id"`

	Config    cluster.Config      `json:"config"`
	Stats     *cluster.Statistics `json:"stats"`
	State     Status              `json:"status"`
	Mode      cluster.OnCrash     `json:"on-crash"`
	StartTime time.Time           `json:"start-time"`

	group     cluster.Cluster
	ETChannel *channel.ManagedChannel
	TLChannel *channel.ManagedChannel

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex
}
