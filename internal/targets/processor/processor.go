package processor

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/buffers"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging/text_logging"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/socket/json_socket"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/terminal"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/component/provision"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/thread"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/thread/provisioner"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/thread/socket"
	provisioner2 "github.com/GabeCordo/FunctionScheduler/internal/targets/processor/use_cases/provisioner"
	socket2 "github.com/GabeCordo/FunctionScheduler/internal/targets/processor/use_cases/socket"
)

const defaultInterruptChannelSize = 1
const defaultChannelSize = 10

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

type Processor struct {
	config  *Config
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
	logger logging.Logger
}

func New(config *Config, repository *ScalingFunctions.Repository) *Processor {

	// check arguments passed to the New function
	if config == nil {
		panic("the *processor.Config argument cannot be nil")
	}

	if repository == nil {
		panic("the *ScalingFunctions.Repository argument cannot be nil")
	}

	// constructor for the *Processor structure
	instance := new(Processor)
	if instance == nil {
		panic("failed to allocate memory for Processor structure")
	}
	instance.config = config

	instance.channels.interrupt = make(chan thread.InterruptEvent, defaultInterruptChannelSize)
	instance.channels.c0 = make(chan *thread.SocketRequest, defaultChannelSize)
	instance.channels.c1 = make(chan *thread.ProvisionerRequest, defaultChannelSize)
	instance.channels.c2 = make(chan *thread.ProvisionerResponse, defaultChannelSize)

	socketConfig := &socket.Config{}
	instance.config.FillSocketConfig(socketConfig)
	socketLogger, err := text_logging.New(Socket.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		panic("failed to initialize a logger for the socket thread")
	}

	socketClient := json_socket.NewClient()

	socketUseCases := socket2.UseCases{Sock: socketClient, Logger: socketLogger}
	instance.threads.socket, err = socket.NewThread(socketConfig, socketLogger, &socketUseCases,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)

	provisionerConfig := &provisioner.Config{}
	instance.config.FillProvisionerConfig(provisionerConfig)
	provisionerLogger, err := text_logging.New(Provisioner.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		panic("failed to initialize a logger for the provisioner thread")
	}

	provisionerBuffer, err := buffers.NewRingBuffer(100) // TODO: this should be a constant
	if err != nil {
		panic("failed to initialize a ring buffer for the provisioner thread")
	}

	provisionerInstance := provision.New(repository, provisionerBuffer)
	provisionerUseCases := provisioner2.UseCases{
		Repository:  repository,
		Provisioner: provisionerInstance,
		Logger:      provisionerLogger,
		Backlog:     provisionerBuffer,
	}

	instance.threads.provisioner, err = provisioner.NewThread(provisionerConfig, &provisionerUseCases,
		instance.channels.interrupt, instance.channels.c0, instance.channels.c1, instance.channels.c2)
	if err != nil {
		panic("failed to instantiate the provisioner thread")
	}

	processorLogger, err := text_logging.New(Undefined.ToString(), &instance.config.Processor.Debug)
	if err != nil {
		panic("failed to initialize a logger for the processor thread")
	}
	instance.logger = processorLogger

	return instance
}

// Connect
// establish a connection with the FunctionScheduler core processes. Upon establishing
// connection, the processor shall register modules to the core and wait
// for processing requests from the core.
func (p *Processor) Connect() {

	p.logger.SetColour(terminal.Purple)

	startingTimestamp := time.Now()

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
		p.logger.Println("system sent SIGTERM or SIGINT signal")
		p.channels.interrupt <- thread.Panic
	case interrupt := <-p.channels.interrupt:
		switch interrupt {
		case thread.Panic:
			p.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			p.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	p.logger.SetColour(terminal.Red)

	p.threads.provisioner.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("modules thread shutdown")
	}

	p.threads.socket.Teardown()
	if p.config.Processor.Debug {
		p.logger.Println("socket processor thread shutdown")
	}

}
