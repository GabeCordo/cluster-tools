package cluster

import (
	"github.com/GabeCordo/etl/components/channel"
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
	ExtractFunc(c *channel.ManagedChannel)
	TransformFunc(in channel.Message) (out channel.Message)
	LoadFunc(in channel.Message)
}

type Config struct {
	Identifier                  string  `json:"identifier"`
	Mode                        OnCrash `json:"on-crash"`
	StartWithNTransformClusters int     `json:"start-with-n-t-channels"`
	StartWithNLoadClusters      int     `json:"start-with-n-l-channels"`
	ETChannelThreshold          int     `json:"et-channel-threshold"`
	ETChannelGrowthFactor       int     `json:"et-channel-growth-factor"`
	TLChannelThreshold          int     `json:"tl-channel-threshold"`
	TLChannelGrowthFactor       int     `json:"tl-channel-growth-factor"`
}

type Statistics struct {
	NumProvisionedExtractRoutines int `json:"num-provisioned-extract-routines"`
	NumProvisionedTransformRoutes int `json:"num-provisioned-transform-routes"`
	NumProvisionedLoadRoutines    int `json:"num-provisioned-load-routines"`
	NumEtThresholdBreaches        int `json:"num-et-threshold-breaches"`
	NumTlThresholdBreaches        int `json:"num-tl-threshold-breaches"`
	NumOfDataProcessed            int `json:"num-of-data-units-processed"`
}

type Status uint8

const (
	Registered = iota
	UnMounted
	Mounted
	InUse
	MarkedForDeletion
)

type Event uint8

const (
	Register = iota
	Mount
	UnMount
	Use
	Delete
)

type Response struct {
	Config     Config        `json:"config"`
	Stats      *Statistics   `json:"stats""`
	LapsedTime time.Duration `json:"lapsed-time"`
	DidItCrash bool          `json:"crashed"`
}
