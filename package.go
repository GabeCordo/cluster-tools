package cluster_tools

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	provisioner_cmp "github.com/GabeCordo/cluster-tools/internal/processor/provisioner"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/http"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/provisioner"
	"github.com/GabeCordo/toolchain/logging"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type State uint8

const (
	Standalone State = iota
	Connected
	Disconnected
)

type Thread uint8

const (
	HttpProcessor Thread = iota
	Provisioner
	Undefined
)

func (module Thread) ToString() string {
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

func NewConfig(name string) *Config {
	config := new(Config)
	config.Name = name
	config.Net.External.Host = "localhost"
	config.Net.External.Port = 5023
	config.Net.Internal.Host = "localhost"
	config.Net.Internal.Port = 5023
	config.MaxAttemptsToCore = 10
	config.Core = "http://localhost:8137"
	config.StandaloneMode = true
	config.StatsMode = true
	config.ReplMode = false
	config.Timeout = 2.0
	config.Debug = true
	return config
}

func (config Config) FillHttpConfig(to *http.Config) {
	to.Debug = &config.Debug
	to.Timeout = &config.Timeout
	to.Standalone = &config.StandaloneMode
	to.Core = &config.Core
	to.ExternalNet = processor.Config{Host: config.Net.External.Host, Port: config.Net.External.Port}
	to.Net = fmt.Sprintf("%s:%d", config.Net.Internal.Host, config.Net.Internal.Port)
}

func (config Config) FillProvisionerConfig(to *provisioner.Config) {
	to.Debug = &config.Debug
	to.Timeout = &config.Timeout
	to.Standalone = &config.StandaloneMode
	to.Core = &config.Core
	to.Processor = processor.Config{Host: config.Net.External.Host, Port: config.Net.External.Port}
}

type Function struct {
	Name  string
	value any
}

type Module struct {
	Name    string
	Version string

	functions map[string]Function

	mutex sync.RWMutex
}

func (module *Module) LinkFunction(name string, value any) error {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.functions[name]; found {
		return errors.New("module already has a function with this name")
	} else {
		module.functions[name] = Function{name, value}
	}

	return nil
}

type Processor struct {
	state State

	threads struct {
		http        *http.Thread
		provisioner *provisioner.Thread
	}

	channels struct {
		interrupt chan threads.InterruptEvent
		c1        chan threads.ProvisionerRequest
		c2        chan threads.ProvisionerResponse
	}

	provisioner *provisioner_cmp.Provisioner

	config *Config
	logger *logging.Logger

	modules map[string]*Module
	mutex   sync.RWMutex
}

func New(cfg ...*Config) (*Processor, error) {
	instance := new(Processor)

	if len(cfg) == 0 {
		instance.config = NewConfig("temp")
	} else if cfg[0] != nil {
		instance.config = cfg[0]
	} else {
		panic(errors.New("the pipeline passed to processor.New cannot be nil"))
	}

	instance.channels.interrupt = make(chan threads.InterruptEvent, 1)
	instance.channels.c1 = make(chan threads.ProvisionerRequest, 10)
	instance.channels.c2 = make(chan threads.ProvisionerResponse, 10)

	httpConfig := &http.Config{}
	instance.config.FillHttpConfig(httpConfig)
	httpLogger, err := logging.NewLogger(HttpProcessor.ToString(), &instance.config.Debug)
	if err != nil {
		return nil, err
	}
	instance.threads.http, err = http.NewThread(httpConfig, httpLogger,
		instance.channels.interrupt, instance.channels.c1, instance.channels.c2)

	provisionerConfig := &provisioner.Config{}
	instance.config.FillProvisionerConfig(provisionerConfig)
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &instance.config.Debug)
	if err != nil {
		return nil, err
	}

	instance.provisioner = provisioner_cmp.New()

	instance.threads.provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger, instance.provisioner,
		instance.channels.interrupt, instance.channels.c1, instance.channels.c2)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Undefined.ToString(), &instance.config.Debug)
	if err != nil {
		return nil, err
	}
	instance.logger = processorLogger

	instance.modules = make(map[string]*Module)

	return instance, nil
}

// Module
// Find or create a new module to encapsulate clusters within. A module can be
// described as a set of clusters that relate in terms of functionality.
func (p *Processor) Module(name string) *Module {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, found := p.modules[name]; !found {

		mod := new(Module)
		mod.Name = name
		mod.functions = make(map[string]Function)

		p.modules[name] = mod
	}

	return p.modules[name]
}

// Run
// Start the processor and wait for SYSINT blocking the calling thread.
func (p *Processor) Run() {

	p.logger.SetColour(logging.Purple)

	// add all the modules to the provisioner
	// note: this is a workaround to avoid exposing too much to the developer
	// TODO: enhanced the comment
	for _, module := range p.modules {
		mod, err := p.provisioner.AddModule(module.Name)
		if err != nil {
			panic(err)
		}

		for _, function := range module.functions {

			err = mod.AddFunction(function.Name, function.value)
			if err != nil {
				panic(err)
			}
		}
	}

	p.threads.provisioner.Setup()
	if p.config.Debug {
		p.logger.Println("started modules thread")
	}
	go p.threads.provisioner.Start()

	p.threads.http.Setup()
	if p.config.Debug {
		p.logger.Println("started http processor thread")
	}
	go p.threads.http.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigs:
		fmt.Println("system sent SIGTERM or SIGINT signal")
		p.channels.interrupt <- threads.Panic
	case interrupt := <-p.channels.interrupt:
		switch interrupt {
		case threads.Panic:
			p.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			p.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	p.logger.SetColour(logging.Red)

	p.threads.http.Teardown()
	if p.config.Debug {
		p.logger.Println("http processor thread shutdown")
	}

	p.threads.provisioner.Teardown()
	if p.config.Debug {
		p.logger.Println("modules thread shutdown")
	}
}

func (p *Processor) Connect(host string) error {

	if p.state == Connected {
		return nil
	}

	p.config.Core = host

	cfg := &processor.Config{Host: p.config.Net.External.Host, Port: p.config.Net.External.Port}

	// attempt to connect to the core 10 times before crashing the processor
	for i := 0; i < p.config.MaxAttemptsToCore; i++ {

		err := api.ConnectToCore(p.config.Core, cfg)
		if err == nil {
			p.logger.Printf("connected to a new core at %s\n", p.config.Core)
			break
		} else {
			p.logger.Alertf("failed to connect to the core at %s\n", p.config.Core)
			if i == (p.config.MaxAttemptsToCore - 1) {
				return errors.New("max attempts to connect to core exceeded")
			} else {
				p.logger.Alertf("attempt to connect attempt %d...\n", i+1)
			}
		}
		time.Sleep(1 * time.Second)
	}

	*p.threads.provisioner.Config.Standalone = false // legacy; todo rework
	p.state = Connected
	return nil
}

func (p *Processor) Disconnect() error {

	if p.state != Connected {
		return errors.New("not connected to a core")
	}

	cfg := &processor.Config{Host: p.config.Net.External.Host, Port: p.config.Net.External.Port}

	defer func() {
		err := api.DisconnectFromCore(p.config.Core, cfg)
		if err == nil {
			p.logger.Printf("disconnected from the core at %s\n", p.config.Core)
		} else {
			p.logger.Alertf("failed to disconnect from the core at %s\n", p.config.Core)
			p.logger.Alertln("\t1. the core is unreachable at the moment")
			p.logger.Alertln("\t2. the core has crashed")
			os.Exit(-1)
		}
	}()

	p.state = Disconnected
	return nil
}

func (p *Processor) Debug(debug bool) {

	p.config.Debug = debug
}
