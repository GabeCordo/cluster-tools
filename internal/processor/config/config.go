package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/http"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/provisioner"
	"os"
)

type NetworkConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type RunConfig struct {
	Name      string   `yaml:"Name"`
	Pipelines []string `yaml:"Pipelines"`
}

type Config struct {
	Processor struct {
		Name           string `yaml:"name" toml:"Name"`
		Debug          bool   `yaml:"debug" toml:"Debug"`
		StandaloneMode bool   `yaml:"standalone" toml:"Standalone,omitempty"`
		Pipeline       struct {
			Default string `yaml:"default,omitempty" toml:"Default,omitempty"`
		} `yaml:"pipeline" toml:"Pipeline"`
		Run     []RunConfig `yaml:"run" toml:"Run"`
		Threads struct {
			Timeout float64 `yaml:"timeout" toml:"Timeout"`
		} `yaml:"threads" toml:"Threads"`
	} `yaml:"processor" toml:"Processor"`
	Core struct {
		Host     string `yaml:"host" toml:"Host"`
		Attempts int    `yaml:"attempts" toml:"Attempts"`
	} `yaml:"core" toml:"Core"`
	Net struct {
		External NetworkConfig `yaml:"external" toml:"External"`
		Internal NetworkConfig `yaml:"internal" toml:"Internal"`
	} `yaml:"net" toml:"Net"`
}

func Load(path string) (*Config, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	_, err = toml.NewDecoder(f).Decode(config)

	if err != nil {
		return nil, err
	} else {
		return config, nil
	}
}
func NewConfig(name string) *Config {
	config := new(Config)
	config.Processor.Name = name
	config.Processor.StandaloneMode = true
	config.Processor.Debug = true
	config.Core.Host = "http://localhost:8137"
	config.Core.Attempts = 10
	config.Net.External.Host = "localhost"
	config.Net.External.Port = 5023
	config.Net.Internal.Host = "localhost"
	config.Net.Internal.Port = 5023
	config.Processor.Threads.Timeout = 2.0
	return config
}

func (config *Config) FillHttpConfig(to *http.Config) {
	to.Debug = &config.Processor.Debug
	to.Timeout = &config.Processor.Threads.Timeout
	to.Standalone = &config.Processor.StandaloneMode
	to.Core = &config.Core.Host
	to.ExternalNet = processor.Config{Host: config.Net.External.Host, Port: config.Net.External.Port}
	to.Net = fmt.Sprintf("%s:%d", config.Net.Internal.Host, config.Net.Internal.Port)
}

func (config *Config) FillProvisionerConfig(to *provisioner.Config) {
	to.Debug = &config.Processor.Debug
	to.Timeout = &config.Processor.Threads.Timeout
	to.Standalone = &config.Processor.StandaloneMode
	to.Core = &config.Core.Host
	to.Processor = processor.Config{Host: config.Net.External.Host, Port: config.Net.External.Port}
}
