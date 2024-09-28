package pipeline

import (
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel/duplex"
	"sync"
)

type Channel struct {
	Identifier string
	Producers  []*Function              `json:"-"`
	Receiver   []*Function              `json:"-"`
	Stats      *statistic.PipeStatistic `json:"-"`
	Value      *duplex.ManagedChannel

	Config struct {
		GrowthFactor float64
		Threshold    int
	} `json:"-"`

	Mutex sync.RWMutex `json:"-"`
}

type Function struct {
	Identifier string
	To         *Channel
	From       *Channel
	Stats      *statistic.FunctionStatistic `json:"-"`
	Quit       []chan bool                  `json:"-"`
	Value      any                          `json:"-"`

	Config struct {
		StartWith  int
		WaitBefore bool
	} `json:"-"`

	Mutex sync.RWMutex `json:"-"`
}

type Pipeline struct {
	Identifier string

	Roots []*Function `json:"-"`
	Tails []*Function

	Channels  []*Channel  `json:"-"`
	Functions []*Function `json:"-"`

	OnStartup  *Function
	OnTeardown *Function

	Stats *statistic.Statistics

	Mutex sync.RWMutex `json:"-"`
}

func New(config *pipeline.Pipeline, functions []any) *Pipeline {

	instance := new(Pipeline)
	instance.Identifier = config.Identifier

	// todo : this mem allocation should not be here
	instance.Stats = statistic.NewStatistics(len(config.Functions), len(config.Pipes))

	instance.Channels = make([]*Channel, len(config.Pipes))
	for i, c := range config.Pipes {

		channel := new(Channel)
		channel.Identifier = c.Identifier
		channel.Config.Threshold = c.Threshold
		channel.Config.GrowthFactor = c.GrowthFactor
		channel.Stats = &instance.Stats.Pipes[i]
		channel.Value = duplex.New(channel.Identifier, c.Threshold, c.GrowthFactor, &instance.Stats.Pipes[i].Timing)
		channel.Receiver = make([]*Function, 0)
		channel.Producers = make([]*Function, 0)

		instance.Channels[i] = channel
	}

	instance.Roots = make([]*Function, 0)
	instance.Tails = make([]*Function, 0)
	instance.Functions = make([]*Function, len(config.Functions))
	for i, f := range config.Functions {

		function := new(Function)
		function.Identifier = f.Identifier
		function.Config.StartWith = f.StartWith
		function.Config.WaitBefore = f.WaitBefore
		function.Stats = &instance.Stats.Functions[i]
		function.Quit = make([]chan bool, 0)
		function.Value = functions[i]

		function.To = nil
		function.From = nil
		for _, c := range instance.Channels {
			if c.Identifier == f.To {
				c.Producers = append(c.Producers, function)
				function.To = c
			}

			if c.Identifier == f.From {
				c.Receiver = append(c.Receiver, function)
				function.From = c
			}
		}

		if function.To == nil {
			instance.Tails = append(instance.Tails, function)
		}

		if function.From == nil {
			instance.Roots = append(instance.Tails, function)
		}

		if function.Identifier == config.OnStartup {
			instance.OnStartup = function
		}

		if function.Identifier == config.OnTeardown {
			instance.OnTeardown = function
		}

		instance.Functions[i] = function
	}

	return instance
}
