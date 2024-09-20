package interfaces

import (
	"time"
)

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

func NewTimingStatistics() *TimingStatistics {
	timing := new(TimingStatistics)
	timing.MinTimeBeforePop = 0
	timing.MaxTimeBeforePop = 0
	timing.AverageTime = 0
	timing.MedianTime = 0
	return timing
}

type Statistics struct {
	Threads struct {
		NumProvisionedExtractRoutines int `json:"num-provisioned-extract-routines"`
		NumActiveExtractRoutines      int `json:"num-active-extract-routines"`
		NumProvisionedTransformRoutes int `json:"num-provisioned-transform-routes"`
		NumActiveTransformRoutines    int `json:"num-active-transform-routines"`
		NumProvisionedLoadRoutines    int `json:"num-provisioned-load-routines"`
		NumActiveLoadRoutines         int `json:"num-active-load-routines"`
	} `json:"threads"`
	Channels struct {
		NumEtThresholdBreaches int `json:"num-et-threshold-breaches"`
		NumTlThresholdBreaches int `json:"num-tl-threshold-breaches"`
	} `json:"channels"`
	Data struct {
		TotalProcessed            int `json:"total-processed"`
		TotalOverETChannel        int `json:"total-over-et"`
		TotalInvalidOverETChannel int `json:"total-invalid-over-et-channel"`
		TotalOverTLChannel        int `json:"total-over-tl"`
		TotalInvalidOverTLChannel int `json:"total-invalid-over-tl-channel"`
		TotalDropped              int `json:"total-dropped"`
	} `json:"data"`
	Timing struct {
		ET               *TimingStatistics `json:"et-channel"`
		etSet            bool
		TL               *TimingStatistics `json:"tl-channel"`
		tlSet            bool
		MaxTotalTime     time.Duration `json:"max-total-time-ns"`
		MinTotalTime     time.Duration `json:"min-total-time-ns"`
		AverageTotalTime time.Duration `json:"avg-total-time-ns"`
		MedianTotalTime  time.Duration `json:"med-total-time-ns"`
		totalSet         bool
	} `json:"timing"`
}

func NewStatistics(et, tl *TimingStatistics) *Statistics {
	stats := new(Statistics)

	stats.Threads.NumProvisionedTransformRoutes = 0
	stats.Threads.NumProvisionedLoadRoutines = 0
	stats.Channels.NumTlThresholdBreaches = 0
	stats.Channels.NumEtThresholdBreaches = 0
	stats.Data.TotalProcessed = 0
	stats.Timing.ET = et
	stats.Timing.TL = tl

	return stats
}

func (statistics *Statistics) CalculateTiming(et, tl *TimingStatistics) {

	statistics.Timing.ET = et
	statistics.Timing.etSet = true
	statistics.Timing.TL = tl
	statistics.Timing.tlSet = true

	if statistics.Timing.ET.MaxTimeBeforePop > statistics.Timing.TL.MaxTimeBeforePop {
		statistics.Timing.MaxTotalTime = statistics.Timing.ET.MaxTimeBeforePop
	} else {
		statistics.Timing.MaxTotalTime = statistics.Timing.TL.MaxTimeBeforePop
	}

	// determine the maximum time between data sitting on the ET and TL channels
	if statistics.Timing.ET.MaxTimeBeforePop > statistics.Timing.TL.MaxTimeBeforePop {
		statistics.Timing.MaxTotalTime = statistics.Timing.ET.MaxTimeBeforePop
	} else {
		statistics.Timing.MaxTotalTime = statistics.Timing.TL.MaxTimeBeforePop
	}

	// determine the minimum time between data sitting on the ET and TL channels
	if statistics.Timing.ET.MinTimeBeforePop < statistics.Timing.TL.MinTimeBeforePop {
		statistics.Timing.MinTotalTime = statistics.Timing.ET.MinTimeBeforePop
	} else {
		statistics.Timing.MinTotalTime = statistics.Timing.TL.MinTimeBeforePop
	}

	statistics.Timing.AverageTotalTime =
		(statistics.Timing.ET.AverageTime + statistics.Timing.TL.AverageTime) / 2

	statistics.Timing.MedianTotalTime =
		(statistics.Timing.ET.MedianTime + statistics.Timing.TL.MedianTime) / 2
}
