package processor

import (
	"errors"
	"fmt"
	cluster "github.com/GabeCordo/clarence/cluster"
	"github.com/GabeCordo/clarence/internal/api"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"github.com/GabeCordo/clarence/internal/threads/common"
	"github.com/GabeCordo/clarence/internal/threads/provisioner"
	"github.com/GabeCordo/clarence/wrapper"
	"github.com/GabeCordo/toolchain/logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run
// Start the processor and wait for SYSINT blocking the calling thread.
func (processor *Processor) Run() {

	var moduleName, clusterName string
	var metadata map[string]string
	var paramParsingErr error

	if processor.state == Standalone {
		moduleName, clusterName, metadata, paramParsingErr = parse()
	}

	processor.Logger.SetColour(logging.Purple)

	if processor.Config.Debug {
		if processor.state == Standalone {
			processor.Logger.Println("running in STANDALONE mode")
		} else {
			processor.Logger.Println("running in CONNECTED mode")
		}
	}

	processor.Provisioner.Setup()
	if processor.Config.Debug {
		processor.Logger.Println("started modules thread")
	}
	go processor.Provisioner.Start()

	processor.HttpThread.Setup()
	if processor.Config.Debug {
		processor.Logger.Println("started http processor thread")
	}
	go processor.HttpThread.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	if (processor.state == Standalone) && (paramParsingErr == nil) {
		processor.C1 <- common.ProvisionerRequest{
			Action:   common.ProvisionerSupervisorCreate,
			Module:   moduleName,
			Cluster:  clusterName,
			Metadata: metadata,
			Config:   &cluster.DefaultConfig,
			Nonce:    0,
		}
	} else if (processor.state == Standalone) && (paramParsingErr != nil) && (processor.Provisioner.NumOfActiveSupervisors() == 0) {
		processor.Interrupt <- common.Shutdown
	}

	select {
	case <-sigs:
		fmt.Println("system sent SIGTERM or SIGINT signal")
		processor.Interrupt <- common.Panic
	case interrupt := <-processor.Interrupt:
		switch interrupt {
		case common.Panic:
			processor.Logger.Printf("[IO] %s\n", " encountered panic")
		default: // shutdown
			processor.Logger.Printf("[IO] %s\n", " shutting down")
		}
	}

	processor.Logger.SetColour(logging.Red)

	processor.HttpThread.Teardown()
	if processor.Config.Debug {
		processor.Logger.Println("http processor thread shutdown")
	}

	processor.Provisioner.Teardown()
	if processor.Config.Debug {
		processor.Logger.Println("modules thread shutdown")
	}
}

func (processor *Processor) Connect(host string) error {

	if processor.state == Connected {
		return nil
	}

	processor.Config.Core = host

	cfg := &interfaces.ProcessorConfig{Host: processor.Config.Net.External.Host, Port: processor.Config.Net.External.Port}

	// attempt to connect to the core 10 times before crashing the processor
	for i := 0; i < processor.Config.MaxAttemptsToCore; i++ {

		err := api.ConnectToCore(processor.Config.Core, cfg)
		if err == nil {
			processor.Logger.Printf("connected to a new core at %s\n", processor.Config.Core)
			break
		} else {
			processor.Logger.Alertf("failed to connect to the core at %s\n", processor.Config.Core)
			if i == (processor.Config.MaxAttemptsToCore - 1) {
				return errors.New("max attempts to connect to core exceeded")
			} else {
				processor.Logger.Alertf("attempt to connect attempt %d...\n", i+1)
			}
		}
		time.Sleep(1 * time.Second)
	}

	processor.Provisioner.Config.Standalone = false // legacy; todo rework
	processor.state = Connected
	return nil
}

func (processor *Processor) Disconnect() error {

	if processor.state != Connected {
		return errors.New("not connected to a core")
	}

	cfg := &interfaces.ProcessorConfig{Host: processor.Config.Net.External.Host, Port: processor.Config.Net.External.Port}

	defer func() {
		err := api.DisconnectFromCore(processor.Config.Core, cfg)
		if err == nil {
			processor.Logger.Printf("disconnected from the core at %s\n", processor.Config.Core)
		} else {
			processor.Logger.Alertf("failed to disconnect from the core at %s\n", processor.Config.Core)
			processor.Logger.Alertln("\t1. the core is unreachable at the moment")
			processor.Logger.Alertln("\t2. the core has crashed")
			os.Exit(-1)
		}
	}()

	processor.state = Disconnected
	return nil
}

func (processor *Processor) Debug(debug bool) {

	processor.Config.Debug = debug
	processor.Provisioner.Config.Debug = debug
	processor.HttpThread.Config.Debug = debug
}

// Module
// Find or create a new module to encapsulate clusters within. A module can be
// described as a set of clusters that relate in terms of functionality.
func (processor *Processor) Module(name string) *wrapper.Module {

	if _, found := provisioner.GetProvisionerInstance().GetModule(name); !found {
		provisioner.GetProvisionerInstance().AddModule(name)
	}

	mod, _ := provisioner.GetProvisionerInstance().GetModule(name)
	return mod
}
