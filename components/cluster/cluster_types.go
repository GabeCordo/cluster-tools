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

type Status uint8

const (
	Registered = iota
	Mounted
	MarkedForDeletion
)

type Response struct {
	Config     Config        `json:"config"`
	Stats      *Statistics   `json:"stats""`
	LapsedTime time.Duration `json:"lapsed-time"`
}
