package pops

import (
	"github.com/Sentmint/pops/internal/processor"
	"github.com/Sentmint/pops/internal/processor/config"
	"github.com/Sentmint/yule"
	"log"
)

type NetworkConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (nCfg NetworkConfig) toInternal() config.NetworkConfig {
	internalNCfg := config.NetworkConfig{}
	internalNCfg.Host = nCfg.Host
	internalNCfg.Port = nCfg.Port
	return internalNCfg
}

type RunConfig struct {
	Name      string   `yaml:"Name"`
	Pipelines []string `yaml:"Pipelines"`
}

func (rCfg RunConfig) toInternal() config.RunConfig {
	internalRunCfg := config.RunConfig{}
	internalRunCfg.Name = rCfg.Name
	internalRunCfg.Pipelines = make([]string, len(rCfg.Pipelines))
	for idx, pipeline := range rCfg.Pipelines {
		internalRunCfg.Pipelines[idx] = pipeline
	}
	return internalRunCfg
}

type Config struct {
	Processor struct {
		Name           string `yaml:"name" toml:"Name"`
		Debug          bool   `yaml:"debug" toml:"Debug"`
		StandaloneMode bool   `yaml:"standalone" toml:"Standalone,omitempty"`
		Pipeline       struct {
			Default string `yaml:"default,omitempty" toml:"Default,omitempty"`
		} `yaml:"pipeline" toml:"Pipeline"`
		Run     []RunConfig `yaml:"run" toml:"Runtime"`
		Threads struct {
			Timeout float64 `yaml:"timeout" toml:"Timeout"`
		} `yaml:"threads" toml:"Threads"`
	} `yaml:"processor" toml:"Processor"`
	Core struct {
		Host     string `yaml:"host" toml:"Host"`
		Attempts int    `yaml:"attempts" toml:"Attempts"`
		TLS      struct {
			Certificate string `yaml:"certificate" toml:"Certificate"`
		} `yaml:"tls" toml:"TLS"`
	} `yaml:"core" toml:"Core"`
}

func (cfg Config) toInternal() *config.Config {
	internalCfg := new(config.Config)
	internalCfg.Processor.Name = cfg.Processor.Name
	internalCfg.Processor.Debug = cfg.Processor.Debug
	internalCfg.Processor.StandaloneMode = cfg.Processor.StandaloneMode
	internalCfg.Processor.Pipeline = cfg.Processor.Pipeline
	internalCfg.Processor.Run = make([]config.RunConfig, len(cfg.Processor.Run))
	for idx, run := range cfg.Processor.Run {
		internalCfg.Processor.Run[idx] = run.toInternal()
	}
	internalCfg.Processor.Threads.Timeout = cfg.Processor.Threads.Timeout
	internalCfg.Core.Host = cfg.Core.Host
	internalCfg.Core.Attempts = cfg.Core.Attempts
	internalCfg.Core.TLS.Certificate = cfg.Core.TLS.Certificate
	return internalCfg
}

// Run
// can be used to attach the processor to the core.
//
// Variants:
// Run( yule.RunnablePipeline ) -> runs in Singular mode
// Run( yule.Repository ) -> runs in AdHoc mode
func Run(config *Config, inputs ...any) error {

	if config == nil {
		log.Fatal("call to Connect() received a nil config")
	}

	cfg := processor.BuilderConfig{}
	cfg.Injectables = make([]any, 0)

	internalCfg := config.toInternal()
	cfg.Config = internalCfg

	for _, input := range inputs {
		if runnable, ok := (input).(*yule.RunnablePipeline); ok {
			// (1) Processor goal: compute and share
			cfg.Runnable = runnable
		} else if repository, ok := (input).(*yule.Repository); ok {
			// (2) Processor goal: listen and compute
			cfg.Repository = repository
		} else {
			cfg.Injectables = append(cfg.Injectables, input)
		}
	}

	p, err := processor.New(cfg)
	if err != nil {
		return err
	}

	return p.Runtime()
}
