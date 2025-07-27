package cluster_tools

import (
	"errors"
	"fmt"
	socket2 "github.com/GabeCordo/Flock/internal/processor/use_cases/socket"
	"github.com/GabeCordo/Flock/internal/shared/socket/json_socket"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/GabeCordo/Flock/internal/processor"
	provisionerCmp "github.com/GabeCordo/Flock/internal/processor/component/provision"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	"github.com/GabeCordo/Flock/internal/processor/thread/provisioner"
	"github.com/GabeCordo/Flock/internal/processor/thread/socket"
	"github.com/GabeCordo/toolchain/logging"
)

type Thread uint8

const (
	Socket Thread = iota
	Provisioner
	Undefined
)

func (module Thread) ToString() string {
	switch module {
	case Socket:
		return "socket"
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

func (module *Module) Map(mappings map[string]any) error {

	for name, value := range mappings {
		if err := module.LinkFunction(name, value); err != nil {
			return err
		}
	}
	return nil
}

type Processor struct {
	threads struct {
		socket      *socket.Thread
		provisioner *provisioner.Thread
	}

	channels struct {
		interrupt chan thread.InterruptEvent
		c0        chan *thread.SocketRequest
		c1        chan *thread.ProvisionerRequest
		c2        chan *thread.ProvisionerResponse
	}

	provisioner *provisionerCmp.Provisioner

	config *processor.Config
	logger *logging.Logger

	modules map[string]*Module
	mutex   sync.RWMutex
}

func New() (*Processor, error) {
	instance := new(Processor)

	var err error

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	workingDir := filepath.Dir(ex)

	// the executable is in the same folder as the config file
	fp := filepath.Join(workingDir, "processor.toml")
	instance.config, err = processor.Load(fp)

	// the executable is in the /bin or /cmd folder
	fp = filepath.Join(workingDir, "..", "processor.toml")
	if err != nil {
		instance.config, err = processor.Load(fp)
	}

	// the executable is in the /cmd/binary-name folder
	fp = filepath.Join(workingDir, "..", "..", "processor.toml")
	if err != nil {
		instance.config, err = processor.Load(fp)
	}

	if err != nil {
		instance.config = processor.NewConfig("debug")
	}

	instance.config.Processor.StandaloneMode = true

	instance.channels.interrupt = make(chan thread.InterruptEvent, 1)
	instance.channels.c0 = make(chan *thread.SocketRequest, 10)
	instance.channels.c1 = make(chan *thread.ProvisionerRequest, 10)
	instance.channels.c2 = make(chan *thread.ProvisionerResponse, 10)

	socketConfig := &socket.Config{}
	instance.config.FillSocketConfig(socketConfig)
	socketLogger, err := logging.NewLogger(Socket.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}

	socketClient := json_socket.NewClient()

	socketUseCases := socket2.UseCases{Sock: socketClient, Logger: socketLogger}
	instance.threads.socket, err = socket.NewThread(socketConfig, socketLogger, &socketUseCases,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)

	provisionerConfig := &provisioner.Config{}
	instance.config.FillProvisionerConfig(provisionerConfig)
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}

	instance.provisioner = provisionerCmp.New()

	instance.threads.provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger, instance.provisioner,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)
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

// Runtime
// Start the processor and wait for SYSINT blocking the calling thread.
func (p *Processor) Runtime() {

	p.logger.SetColour(logging.Purple)

	startingTimestamp := time.Now()

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

	p.threads.socket.Setup()
	if p.config.Processor.Debug {
		p.logger.Println("started socket processor thread")
	}
	go p.threads.socket.Start()

	doneStartingTimestamp := time.Now()
	timeToStart := doneStartingTimestamp.Sub(startingTimestamp).Milliseconds()
	if p.config.Processor.Debug {
		p.logger.Printf("startup took %d ms\n", timeToStart)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigs:
		fmt.Println("system sent SIGTERM or SIGINT signal")
		p.channels.interrupt <- thread.Panic
	case interrupt := <-p.channels.interrupt:
		switch interrupt {
		case thread.Panic:
			p.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			p.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	p.logger.SetColour(logging.Red)

	p.threads.provisioner.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("modules thread shutdown")
	}

	p.threads.socket.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("socket processor thread shutdown")
	}

}

func (p *Processor) Debug(debug bool) {

	p.config.Processor.Debug = debug
}
