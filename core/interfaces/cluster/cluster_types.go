package cluster

import (
	"github.com/GabeCordo/mango/core/interfaces/channel"
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

type Transformer func(in []any) (out any, success bool)

// M contains metadata about the running supervisor including any state
// information that a devleoper might need to interact with the Supervisor.
type M interface {
	GetKey(key string) string
}

// Cluster is a set of functions that define an ETL process
// cluster functions are provisioned on goroutines to run in parallel and
// process data.
type Cluster interface {
	ExtractFunc(metadata M, c channel.OneWay)
	TransformFunc(metadata M, in any) (out any, success bool)
}

type LoadAll interface {
	LoadFunc(metadata M, in []any)
}

type LoadOne interface {
	LoadFunc(metadata M, in any)
}

type VerifiableET interface {
	VerifyETFunction(in any) (valid bool)
}

type VerifiableTL interface {
	VerifyTLFunction(in any) (valid bool)
}

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

type DataTiming struct {
	ETIn  time.Time
	ETOut time.Time
	TLIn  time.Time
	TLOut time.Time
}

type TimingStatistics struct {
	MinTimeBeforePop time.Duration `json:"min-time-before-pop-ns"`
	MaxTimeBeforePop time.Duration `json:"max-time-before-pop-ns"`
	AverageTime      time.Duration `json:"average-time-ns"`
	MedianTime       time.Duration `json:"median-time-ns"`
}

type Statistics struct {
	Threads struct {
		NumProvisionedExtractRoutines int `json:"num-provisioned-extract-routines"`
		NumProvisionedTransformRoutes int `json:"num-provisioned-transform-routes"`
		NumProvisionedLoadRoutines    int `json:"num-provisioned-load-routines"`
	} `json:"threads"`
	Channels struct {
		NumEtThresholdBreaches int `json:"num-et-threshold-breaches"`
		NumTlThresholdBreaches int `json:"num-tl-threshold-breaches"`
	} `json:"channels"`
	Data struct {
		TotalProcessed     int `json:"total-processed"`
		TotalOverETChannel int `json:"total-over-et"`
		TotalOverTLChannel int `json:"total-over-tl"`
		TotalDropped       int `json:"total-dropped"`
	} `json:"data"`
	Timing struct {
		ET               TimingStatistics `json:"et-channel"`
		etSet            bool
		TL               TimingStatistics `json:"tl-channel"`
		tlSet            bool
		MaxTotalTime     time.Duration `json:"max-total-time-ns"`
		MinTotalTime     time.Duration `json:"min-total-time-ns"`
		AverageTotalTime time.Duration `json:"avg-total-time-ns"`
		MedianTotalTime  time.Duration `json:"med-total-time-ns"`
		totalSet         bool
	} `json:"timing"`
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
