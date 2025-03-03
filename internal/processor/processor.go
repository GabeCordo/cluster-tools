package processor

import (
	"fmt"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/Sentmint/pops/internal/processor/config"
	"github.com/Sentmint/pops/internal/processor/threads"
	"github.com/Sentmint/pops/internal/processor/threads/provisioner"
	"github.com/Sentmint/pops/internal/processor/threads/socket"
	"github.com/Sentmint/yule"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const ClusterToolsConfigEnvVar = "CTOOLS_CONFIG"

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

type BuilderConfig struct {
	Repository  *yule.Repository
	Runnable    *yule.RunnablePipeline
	Injectables []any
	Config      *config.Config
}

type Processor struct {
	threads struct {
		socket      *socket.Thread
		provisioner *provisioner.Thread
	}

	channels struct {
		interrupt chan threads.InterruptEvent
		c0        chan threads.SocketRequest
		c1        chan threads.ProvisionerRequest
		c2        chan threads.ProvisionerResponse
	}

	config *config.Config
	logger *logging.Logger

	mutex sync.RWMutex
}

func New(builderConfig BuilderConfig) (*Processor, error) {
	instance := new(Processor)

	if builderConfig.Config == nil {
		log.Fatal("building a processor cannot use a nil config")
	}

	if (builderConfig.Runnable == nil) && (builderConfig.Repository == nil) {
		log.Fatal("building a processor requires at least a runnable or repository")
	}

	instance.config = builderConfig.Config
	instance.config.Processor.StandaloneMode = true

	instance.channels.interrupt = make(chan threads.InterruptEvent, 1)
	instance.channels.c0 = make(chan threads.SocketRequest, 10)
	instance.channels.c1 = make(chan threads.ProvisionerRequest, 10)
	instance.channels.c2 = make(chan threads.ProvisionerResponse, 10)

	socketConfig := &socket.Config{}
	instance.config.FillSocketConfig(socketConfig)
	httpLogger, err := logging.NewLogger(Socket.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}
	instance.threads.socket, err = socket.NewThread(socketConfig, httpLogger,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)

	provisionerConfig := &provisioner.Config{}
	instance.config.FillProvisionerConfig(provisionerConfig)
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}

	instance.threads.provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger,
		builderConfig.Repository, builderConfig.Runnable, builderConfig.Injectables,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Undefined.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		return nil, err
	}
	instance.logger = processorLogger

	return instance, nil
}

// Runtime
// Start the processor and wait for SYSINT blocking the calling thread.
func (p *Processor) Runtime() error {

	p.logger.SetColour(logging.Purple)

	startingTimestamp := time.Now()

	// TODO: do we need this?
	// add all the modules to the provisioner
	// note: this is a workaround to avoid exposing too much to the developer
	// TODO: enhanced the comment
	//for _, module := range p.modules {
	//	mod, err := p.provisioner.AddModule(module.Name)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	for _, function := range module.functions {
	//
	//		err = mod.AddFunction(function.Name, function.value)
	//		if err != nil {
	//			panic(err)
	//		}
	//	}
	//}

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

	p.threads.provisioner.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("modules thread shutdown")
	}

	p.threads.socket.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("socket processor thread shutdown")
	}

	return nil
}
