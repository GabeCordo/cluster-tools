package processor

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"github.com/GabeCordo/clarence/internal/threads/common"
	"github.com/GabeCordo/clarence/internal/threads/http"
	"github.com/GabeCordo/clarence/internal/threads/provisioner"
	"github.com/GabeCordo/toolchain/logging"
)

type State uint8

const (
	Standalone State = iota
	Connected
	Disconnected
)

type Module uint8

const (
	HttpProcessor Module = iota
	Provisioner
	Undefined
)

func (module Module) ToString() string {
	switch module {
	case HttpProcessor:
		return "http-processor"
	case Provisioner:
		return "modules"
	default:
		return "-"
	}
}

type NetworkConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Name              string  `yaml:"name"`
	Debug             bool    `yaml:"debug"`
	StandaloneMode    bool    `yaml:"standalone"`
	ReplMode          bool    `yaml:"repl"`
	StatsMode         bool    `yaml:"stats"`
	Timeout           float64 `yaml:"timeout"`
	Core              string  `yaml:"core"`
	MaxAttemptsToCore int     `yaml:"max_attempts_to_core"`
	Net               struct {
		External NetworkConfig `yaml:"external"`
		Internal NetworkConfig `yaml:"internal"`
	} `yaml:"net"`
}

type Processor struct {
	state State

	HttpThread  *http.Thread
	Provisioner *provisioner.Thread

	Interrupt chan common.InterruptEvent
	C1        chan common.ProvisionerRequest
	C2        chan common.ProvisionerResponse

	Config *Config

	Logger *logging.Logger
}

func New(cfg ...*Config) (*Processor, error) {
	processor := new(Processor)

	if len(cfg) == 0 {
		processor.Config = NewConfig("temp")
	} else if cfg[0] != nil {
		processor.Config = cfg[0]
	} else {
		panic(errors.New("the config passed to processor.New cannot be nil"))
	}

	processor.Interrupt = make(chan common.InterruptEvent, 1)
	processor.C1 = make(chan common.ProvisionerRequest, 10)
	processor.C2 = make(chan common.ProvisionerResponse, 10)

	httpConfig := &http.Config{
		Debug:   processor.Config.Debug,
		Timeout: processor.Config.Timeout,
		Net:     fmt.Sprintf("%s:%d", processor.Config.Net.Internal.Host, processor.Config.Net.Internal.Port),
	}
	httpLogger, err := logging.NewLogger(HttpProcessor.ToString(), &processor.Config.Debug)
	if err != nil {
		return nil, err
	}
	processor.HttpThread, err = http.NewThread(httpConfig, httpLogger,
		processor.Interrupt, processor.C1, processor.C2)

	provisionerConfig := &provisioner.Config{
		Debug:      true,
		Timeout:    processor.Config.Timeout,
		Standalone: processor.Config.StandaloneMode,
		Core:       processor.Config.Core,
		Processor:  interfaces.ProcessorConfig{Host: processor.Config.Net.External.Host, Port: processor.Config.Net.External.Port},
	}
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &processor.Config.Debug)
	if err != nil {
		return nil, err
	}
	processor.Provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger,
		processor.Interrupt, processor.C1, processor.C2)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Undefined.ToString(), &processor.Config.Debug)
	if err != nil {
		return nil, err
	}
	processor.Logger = processorLogger

	return processor, nil
}
