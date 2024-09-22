package cluster_tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/http"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads/provisioner"
	"github.com/GabeCordo/toolchain/logging"
	"os"
	"os/signal"
	"strings"
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
	config.Debug = false
	return config
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

	config *Config
	logger *logging.Logger

	modules map[string]*Module
	mutex   sync.RWMutex
}

func New(cfg ...*Config) (*Processor, error) {
	processor := new(Processor)

	if len(cfg) == 0 {
		processor.config = NewConfig("temp")
	} else if cfg[0] != nil {
		processor.config = cfg[0]
	} else {
		panic(errors.New("the pipeline passed to processor.New cannot be nil"))
	}

	processor.channels.interrupt = make(chan threads.InterruptEvent, 1)
	processor.channels.c1 = make(chan threads.ProvisionerRequest, 10)
	processor.channels.c2 = make(chan threads.ProvisionerResponse, 10)

	httpConfig := &http.Config{
		Debug:   processor.config.Debug,
		Timeout: processor.config.Timeout,
		Net:     fmt.Sprintf("%s:%d", processor.config.Net.Internal.Host, processor.config.Net.Internal.Port),
	}
	httpLogger, err := logging.NewLogger(HttpProcessor.ToString(), &processor.config.Debug)
	if err != nil {
		return nil, err
	}
	processor.threads.http, err = http.NewThread(httpConfig, httpLogger,
		processor.channels.interrupt, processor.channels.c1, processor.channels.c2)

	provisionerConfig := &provisioner.Config{
		Debug:      true,
		Timeout:    processor.config.Timeout,
		Standalone: processor.config.StandaloneMode,
		Core:       processor.config.Core,
		Processor:  interfaces.ProcessorConfig{Host: processor.config.Net.External.Host, Port: processor.config.Net.External.Port},
	}
	provisionerLogger, err := logging.NewLogger(Provisioner.ToString(), &processor.config.Debug)
	if err != nil {
		return nil, err
	}
	processor.threads.provisioner, err = provisioner.NewThread(provisionerConfig, provisionerLogger,
		processor.channels.interrupt, processor.channels.c1, processor.channels.c2)
	if err != nil {
		return nil, err
	}

	processorLogger, err := logging.NewLogger(Undefined.ToString(), &processor.config.Debug)
	if err != nil {
		return nil, err
	}
	processor.logger = processorLogger

	return processor, nil
}

func parse() (module, cluster string, metadata map[string]string, err error) {

	args := os.Args[1:]
	numArgs := len(args)

	if numArgs >= 1 {

		moduleCluster := args[0]
		splitModuleCluster := strings.Split(moduleCluster, ":")

		if len(splitModuleCluster) != 2 {
			return "", "", nil, errors.New("expected first parameter to be in the format module:cluster")
		}

		module = splitModuleCluster[0]
		cluster = splitModuleCluster[1]
	}

	metadata = make(map[string]string)

	if numArgs >= 2 {
		metadataStr := ""

		for i := 1; i < numArgs; i++ {
			metadataStr += args[i]
		}

		fmt.Println(metadataStr)

		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			output := fmt.Sprintf("received metadata is not a valid json: \n%s\n", metadataStr)
			return "", "", nil, errors.New(output)
		}
	}

	return module, cluster, metadata, nil
}

// Run
// Start the processor and wait for SYSINT blocking the calling thread.
func (processor *Processor) Run() {

	var moduleName, clusterName string
	var metadata map[string]string
	var paramParsingErr error

	if processor.state == Standalone {
		moduleName, clusterName, metadata, paramParsingErr = parse()
	}

	processor.logger.SetColour(logging.Purple)

	if processor.config.Debug {
		if processor.state == Standalone {
			processor.logger.Println("running in STANDALONE mode")
		} else {
			processor.logger.Println("running in CONNECTED mode")
		}
	}

	processor.threads.provisioner.Setup()
	if processor.config.Debug {
		processor.logger.Println("started modules thread")
	}
	go processor.threads.provisioner.Start()

	processor.threads.http.Setup()
	if processor.config.Debug {
		processor.logger.Println("started http processor thread")
	}
	go processor.threads.http.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	if (processor.state == Standalone) && (paramParsingErr == nil) {
		processor.channels.c1 <- threads.ProvisionerRequest{
			Action:   threads.ProvisionerSupervisorCreate,
			Module:   moduleName,
			Cluster:  clusterName,
			Metadata: metadata,
			Config:   &cluster.DefaultConfig,
			Nonce:    0,
		}
	} else if (processor.state == Standalone) && (paramParsingErr != nil) && (processor.threads.provisioner.NumOfActiveSupervisors() == 0) {
		processor.channels.interrupt <- threads.Shutdown
	}

	select {
	case <-sigs:
		fmt.Println("system sent SIGTERM or SIGINT signal")
		processor.channels.interrupt <- threads.Panic
	case interrupt := <-processor.channels.interrupt:
		switch interrupt {
		case threads.Panic:
			processor.logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			processor.logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	processor.logger.SetColour(logging.Red)

	processor.threads.http.Teardown()
	if processor.config.Debug {
		processor.logger.Println("http processor thread shutdown")
	}

	processor.threads.provisioner.Teardown()
	if processor.config.Debug {
		processor.logger.Println("modules thread shutdown")
	}
}

func (processor *Processor) Connect(host string) error {

	if processor.state == Connected {
		return nil
	}

	processor.config.Core = host

	cfg := &interfaces.ProcessorConfig{Host: processor.config.Net.External.Host, Port: processor.config.Net.External.Port}

	// attempt to connect to the core 10 times before crashing the processor
	for i := 0; i < processor.config.MaxAttemptsToCore; i++ {

		err := api.ConnectToCore(processor.config.Core, cfg)
		if err == nil {
			processor.logger.Printf("connected to a new core at %s\n", processor.config.Core)
			break
		} else {
			processor.logger.Alertf("failed to connect to the core at %s\n", processor.config.Core)
			if i == (processor.config.MaxAttemptsToCore - 1) {
				return errors.New("max attempts to connect to core exceeded")
			} else {
				processor.logger.Alertf("attempt to connect attempt %d...\n", i+1)
			}
		}
		time.Sleep(1 * time.Second)
	}

	processor.threads.provisioner.Config.Standalone = false // legacy; todo rework
	processor.state = Connected
	return nil
}

func (processor *Processor) Disconnect() error {

	if processor.state != Connected {
		return errors.New("not connected to a core")
	}

	cfg := &interfaces.ProcessorConfig{Host: processor.config.Net.External.Host, Port: processor.config.Net.External.Port}

	defer func() {
		err := api.DisconnectFromCore(processor.config.Core, cfg)
		if err == nil {
			processor.logger.Printf("disconnected from the core at %s\n", processor.config.Core)
		} else {
			processor.logger.Alertf("failed to disconnect from the core at %s\n", processor.config.Core)
			processor.logger.Alertln("\t1. the core is unreachable at the moment")
			processor.logger.Alertln("\t2. the core has crashed")
			os.Exit(-1)
		}
	}()

	processor.state = Disconnected
	return nil
}

func (processor *Processor) Debug(debug bool) {

	processor.config.Debug = debug
	processor.threads.provisioner.Config.Debug = debug
	processor.threads.http.Config.Debug = debug
}

type Function struct {
	Name    string
	Mounted bool
	value   any
}

type Module struct {
	Name    string
	Version string
	Mounted bool

	pipeline map[string]*Function

	mutex sync.RWMutex
}

// Module
// Find or create a new module to encapsulate clusters within. A module can be
// described as a set of clusters that relate in terms of functionality.
func (processor *Processor) Module(name string) *Module {

	processor.mutex.Lock()
	defer processor.mutex.Unlock()

	if _, found := processor.modules[name]; !found {

		mod := new(Module)
		mod.Name = name

		processor.modules[name] = mod
	}

	return processor.modules[name]
}

func Chain(identifier string, functions ...any) *supervisor.Supervisor {

	panic("implement me")
}
