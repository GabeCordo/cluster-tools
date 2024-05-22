package interfaces

import (
	"time"
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type OnCrash string

const (
	Restart   OnCrash = "Restart"
	DoNothing         = "DoNothing"
)

type OnLoad string

const (
	CompleteAndPush OnLoad = "CompleteAndPush"
	WaitAndPush            = "WaitAndPush"
)

type EtlMode string

const (
	Batch  EtlMode = "Batch"
	Stream         = "Stream"
)

type Config struct {
	Identifier                  string  `json:"identifier"`
	OnLoad                      OnLoad  `json:"on-load"`
	OnCrash                     OnCrash `json:"on-crash"`
	StartWithNTransformClusters int     `json:"start-with-n-t-channels"`
	StartWithNLoadClusters      int     `json:"start-with-n-l-channels"`
	ETChannelThreshold          int     `json:"et-channel-threshold"`
	ETChannelGrowthFactor       int     `json:"et-channel-growth-factor"`
	TLChannelThreshold          int     `json:"tl-channel-threshold"`
	TLChannelGrowthFactor       int     `json:"tl-channel-growth-factor"`
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
	Config     Config        `json:"core"`
	Stats      *Statistics   `json:"stats"`
	LapsedTime time.Duration `json:"lapsed-time"`
	DidItCrash bool          `json:"crashed"`
}
