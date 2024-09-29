package cluster_tools

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/cluster-tools/internal/processor/config"
	provisioner_cmp "github.com/GabeCordo/cluster-tools/internal/processor/provision"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/http"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/provisioner"
	"github.com/GabeCordo/toolchain/logging"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const ClusterToolsConfigEnvVar = "CTOOLS_CONFIG"
const ClusterToolsDeploymentsEnvVar = "CTOOLS_DEPLOYMENTS"

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

	config *config.Config
	logger *logging.Logger

	modules map[string]*Module
	mutex   sync.RWMutex
}

func New() (*Processor, error) {
	instance := new(Processor)

	var err error
	if configPath := os.Getenv(ClusterToolsConfigEnvVar); configPath != "" {
		instance.config, err = config.Load(configPath)
		if err != nil {
			panic(err)
		}
	} else {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		workingDir := filepath.Dir(ex)

		// the executable is in the same folder as the config file
		instance.config, err = config.Load(workingDir + "/processor.toml")

		// the executable is in the /bin or /cmd folder
		if err != nil {
			instance.config, err = config.Load(workingDir + "/../processor.toml")
		}

		// the executable is in the /cmd/binary-name folder
		if err != nil {
			instance.config, err = config.Load(workingDir + "/../../processor.toml")
		}

		if err != nil {
			panic("cannot find a processor.toml file in any of the expected directories")
		}
	}

	instance.channels.interrupt = make(chan threads.InterruptEvent, 1)
	instance.channels.c1 = make(chan threads.ProvisionerRequest, 10)
	instance.channels.c2 = make(chan threads.ProvisionerResponse, 10)

	httpConfig := &http.Config{}
	instance.config.FillHttpConfig(httpConfig)
	httpLogger, err := logging.NewLogger(HttpProcessor.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}
	instance.threads.http, err = http.NewThread(httpConfig, httpLogger,
		instance.channels.interrupt, instance.channels.c1, instance.channels.c2)

	provisionerConfig := &provisioner.Config{}
	instance.config.FillProvisionerConfig(provisionerConfig)
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}

	instance.provisioner = provisioner_cmp.New()

	instance.threads.provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger, instance.provisioner,
		instance.channels.interrupt, instance.channels.c1, instance.channels.c2)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Undefined.ToString(), &instance.config.Processor.Debug)
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
	if p.config.Processor.Debug {
		p.logger.Println("started modules thread")
	}
	go p.threads.provisioner.Start()

	p.threads.http.Setup()
	if p.config.Processor.Debug {
		p.logger.Println("started http processor thread")
	}
	go p.threads.http.Start()

	// check if the config has a default pipeline to run on start

	if p.config.Processor.Pipeline.Default != "" {

		deploymentsDir := os.Getenv(ClusterToolsDeploymentsEnvVar)

		var err error
		if deploymentsDir != "" {

			if f, err := os.Stat(ClusterToolsDeploymentsEnvVar); err != nil || !f.IsDir() {
				panic("cannot find deployments directory")
			}

			pipelineFile := ClusterToolsDeploymentsEnvVar + "/" + p.config.Processor.Pipeline.Default + ".yml"
			if _, err := os.Stat(pipelineFile); err != nil {
				panic("no pipeline exists with that default identifier")
			}

			p.config, err = config.Load(pipelineFile)
			if err != nil {
				panic(err)
			}
		} else {

			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			workingDir := filepath.Dir(ex)

			fileName := fmt.Sprintf("/deployments/%s.yml", p.config.Processor.Pipeline.Default)

			// the executable is in the same folder as the config file
			f, err := os.Open(workingDir + fileName)

			// the executable is in the /bin or /cmd folder
			if err != nil {
				f, err = os.Open(workingDir + "/.." + fileName)
			}

			// the executable is in the /cmd/binary-name folder
			if err != nil {
				f, err = os.Open(workingDir + "/../.." + fileName)
			}

			if err != nil {
				panic(fmt.Sprintf("cannot find pipeline %s file in the deployments directory.\n", fileName))
			}

			wrapper := &struct {
				Pipeline *pipeline.Pipeline `yaml:"pipeline"`
			}{}

			if err = yaml.NewDecoder(f).Decode(wrapper); err != nil {
				log.Println("the default pipeline file is corrupted")
			}

			p.channels.c1 <- threads.ProvisionerRequest{
				Action:     threads.ProvisionerRunCreate,
				Namespace:  "common",
				Supervisor: 0,
				Pipeline:   wrapper.Pipeline,
				Metadata:   make(map[string]string),
			}
		}
	}

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
	if p.config.Processor.Debug {
		p.logger.Println("http processor thread shutdown")
	}

	p.threads.provisioner.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("modules thread shutdown")
	}
}

func (p *Processor) Connect(host string) error {

	if p.state == Connected {
		return nil
	}

	p.config.Core.Host = host

	cfg := &processor.Config{Host: p.config.Net.External.Host, Port: p.config.Net.External.Port}

	// attempt to connect to the core 10 times before crashing the processor
	for i := 0; i < p.config.Core.Attempts; i++ {

		err := api.ConnectToCore(p.config.Core.Host, cfg)
		if err == nil {
			p.logger.Printf("connected to a new core at %s\n", p.config.Core)
			break
		} else {
			p.logger.Alertf("failed to connect to the core at %s\n", p.config.Core)
			if i == (p.config.Core.Attempts - 1) {
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
		err := api.DisconnectFromCore(p.config.Core.Host, cfg)
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

	p.config.Processor.Debug = debug
}
