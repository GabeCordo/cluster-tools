package processor

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/GabeCordo/Flock/internal/processor/thread/provisioner"
	"github.com/GabeCordo/Flock/internal/processor/thread/socket"
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
		Run     []RunConfig `yaml:"run" toml:"Runtime"`
		Threads struct {
			Timeout float64 `yaml:"timeout" toml:"Timeout"`
		} `yaml:"thread" toml:"Threads"`
	} `yaml:"processor" toml:"Processor"`
	Core struct {
		Host     string `yaml:"host" toml:"Host"`
		Attempts int    `yaml:"attempts" toml:"Attempts"`
		TLS      struct {
			Certificate string `yaml:"certificate" toml:"Certificate"`
		} `yaml:"tls" toml:"TLS"`
	} `yaml:"core" toml:"Core"`
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
	config.Core.Host = "0.0.0.0:8137"
	config.Core.Attempts = 10
	config.Processor.Threads.Timeout = 2.0
	return config
}

func (config *Config) FillSocketConfig(to *socket.Config) {
	to.Debug = &config.Processor.Debug
	to.Timeout = &config.Processor.Threads.Timeout
	to.Standalone = &config.Processor.StandaloneMode
	to.Core = &config.Core.Host
	to.Tls.Certificate = config.Core.TLS.Certificate
}

func (config *Config) FillProvisionerConfig(to *provisioner.Config) {
	to.Debug = &config.Processor.Debug
	to.Timeout = &config.Processor.Threads.Timeout
	to.Standalone = &config.Processor.StandaloneMode
	to.Core = &config.Core.Host
}
