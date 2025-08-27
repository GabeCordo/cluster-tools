package statistic

import (
	"github.com/FortifiedCode/plover"
	"time"
)

type DataTiming struct {
	ETIn  time.Time
	ETOut time.Time
	TLIn  time.Time
	TLOut time.Time
}

type TimingStatistics struct {
	MinTimeBeforePop time.Duration `json:"min_time_before_pop_ns"`
	MaxTimeBeforePop time.Duration `json:"max_time_before_pop_ns"`
	AverageTime      time.Duration `json:"average_time_ns"`
	MedianTime       time.Duration `json:"median_time_ns"`
}

type FunctionStatistic struct {
	Active     int `json:"active"`
	Provisions int `json:"provisions"`
}

type PipeStatistic struct {
	Pushed   int
	Pulled   int              `json:"processed"`
	Dropped  int              `json:"dropped"`
	Breaches int              `json:"breaches"`
	Timing   TimingStatistics `json:"timing"`
}

type Statistics struct {
	NumOfFunctions int                 `json:"num_of_functions"`
	Functions      []FunctionStatistic `json:"steps"`
	NumOfChannels  int                 `json:"num_of_channels"`
	Pipes          []PipeStatistic     `json:"channels"`
}

func NewStatistics(numOfFunctions, numOfPipes int) *Statistics {
	stats := new(Statistics)

	stats.NumOfFunctions = 0
	stats.Functions = make([]FunctionStatistic, numOfFunctions)

	stats.NumOfChannels = 0
	stats.Pipes = make([]PipeStatistic, numOfPipes)

	return stats
}

type Wrapper struct {
	Timestamp time.Time         `json:"timestamp"`
	Namespace string            `json:"namespace"`
	Pipeline  string            `json:"pipeline"`
	Elapsed   time.Duration     `json:"elapsed"`
	Stats     plover.Statistics `json:"statistics"`
}

type DataTimer struct {
	In  time.Time
	Out time.Time
}

// TODO : move out of here to somewhere ELSE!

func (dataTiming DataTiming) Valid() bool {
	return !dataTiming.ETIn.IsZero() && !dataTiming.ETOut.IsZero() && !dataTiming.TLIn.IsZero() && !dataTiming.TLOut.IsZero()
}

//
//func (statistics *Statistics) CalculateTiming(et, tl map[uint64]DataTimer) {
//
//	// conjoin the statistics from the two channels into one struct
//	timing := make(map[uint64]*DataTiming)
//
//	for id, etTimer := range et {
//		timeStruct := new(DataTiming)
//		timeStruct.ETIn = etTimer.In
//		timeStruct.ETOut = etTimer.Out
//		timing[id] = timeStruct
//	}
//
//	for id, tlTimer := range tl {
//		if timer, found := timing[id]; found {
//			timer.TLIn = tlTimer.In
//			timer.TLOut = tlTimer.Out
//		}
//	}
//
//	totalTimings := make([]time.Duration, 0)
//
//	// calculate the individual max/min timing for each channel
//	for _, timer := range timing {
//
//		if !timer.Valid() {
//			continue
//		}
//
//		timeOnET := timer.ETOut.Sub(timer.ETIn)
//		timeOnTL := timer.TLOut.Sub(timer.TLIn)
//
//		totalTime := timer.TLOut.Sub(timer.ETIn)
//		totalTimings = append(totalTimings, totalTime)
//
//		// set the timing for ET
//		if !statistics.Timing.etSet || (statistics.Timing.ET.MaxTimeBeforePop < timeOnET) {
//			statistics.Timing.ET.MaxTimeBeforePop = timeOnET
//		}
//
//		if !statistics.Timing.etSet || (statistics.Timing.ET.MinTimeBeforePop > timeOnTL) {
//			statistics.Timing.ET.MinTimeBeforePop = timeOnET
//		}
//
//		if !statistics.Timing.etSet {
//			statistics.Timing.etSet = true
//		}
//
//		if !statistics.Timing.tlSet || (statistics.Timing.TL.MaxTimeBeforePop < timeOnTL) {
//			statistics.Timing.TL.MaxTimeBeforePop = timeOnTL
//		}
//
//		if !statistics.Timing.tlSet || (statistics.Timing.TL.MinTimeBeforePop > timeOnTL) {
//			statistics.Timing.TL.MinTimeBeforePop = timeOnTL
//		}
//
//		if !statistics.Timing.tlSet {
//			statistics.Timing.tlSet = true
//		}
//
//		if !statistics.Timing.totalSet {
//			statistics.Timing.MaxTotalTime = totalTime
//			statistics.Timing.MinTotalTime = totalTime
//			statistics.Timing.totalSet = true
//		}
//
//		if statistics.Timing.MaxTotalTime < totalTime {
//			statistics.Timing.MaxTotalTime = totalTime
//		}
//
//		if statistics.Timing.MinTotalTime > totalTime {
//			statistics.Timing.MinTotalTime = totalTime
//		}
//	}
//
//	// order the timings on the et & tl channels into arrays
//	etTimings := make([]time.Duration, 0)
//	tlTimings := make([]time.Duration, 0)
//
//	for _, timing := range et {
//		dur := timing.Out.Sub(timing.In)
//		etTimings = append(etTimings, dur)
//	}
//	sort.Slice(etTimings, func(i, j int) bool { return etTimings[i] < etTimings[j] })
//
//	for _, timing := range tl {
//		dur := timing.Out.Sub(timing.In)
//		tlTimings = append(tlTimings, dur)
//	}
//	sort.Slice(tlTimings, func(i, j int) bool { return tlTimings[i] < tlTimings[j] })
//
//	// calculate the average on the et & tl channels
//	sumEtTimings := time.Duration(0)
//	numEtTimings := int64(len(etTimings))
//	for _, timing := range etTimings {
//		sumEtTimings += timing
//	}
//	statistics.Timing.ET.AverageTime = time.Duration(sumEtTimings.Nanoseconds() / numEtTimings)
//
//	sumTlTimings := time.Duration(0)
//	numTlTimings := int64(len(tlTimings))
//	for _, timing := range tlTimings {
//		sumTlTimings += timing
//	}
//	statistics.Timing.TL.AverageTime = time.Duration(sumTlTimings.Nanoseconds() / numTlTimings)
//
//	sumTotalTimings := time.Duration(0)
//	numTotalTimings := int64(len(totalTimings))
//	for _, timing := range totalTimings {
//		sumTotalTimings += timing
//	}
//	statistics.Timing.AverageTotalTime = time.Duration(sumTotalTimings.Nanoseconds() / numTotalTimings)
//
//	// calculate the median on the et & tl channels
//	etMiddleIndex := numEtTimings / 2
//	if (numEtTimings % 2) == 0 {
//		a := tlTimings[etMiddleIndex-1]
//		b := tlTimings[etMiddleIndex]
//		statistics.Timing.ET.MedianTime = (a + b) / 2
//	} else {
//		statistics.Timing.ET.MedianTime = tlTimings[etMiddleIndex]
//	}
//
//	tlMiddleIndex := numTlTimings / 2
//	if (numTlTimings % 2) == 0 {
//		statistics.Timing.TL.MedianTime = (tlTimings[tlMiddleIndex-1] + tlTimings[tlMiddleIndex]) / 2
//	} else {
//		statistics.Timing.TL.MedianTime = tlTimings[tlMiddleIndex]
//	}
//
//	totalMiddleIndex := numTotalTimings / 2
//	if (numTotalTimings % 2) == 0 {
//		statistics.Timing.MedianTotalTime = (totalTimings[totalMiddleIndex-1] + totalTimings[totalMiddleIndex]) / 2
//	} else {
//		statistics.Timing.MedianTotalTime = totalTimings[totalMiddleIndex]
//	}
//}
