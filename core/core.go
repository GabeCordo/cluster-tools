package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	commonConfigPaths = [...]string{"config.etl.json", "/config/config.etl.json"}
	configLock        = &sync.Mutex{}
	ConfigInstance    *Config
)

func GetConfigInstance() *Config {
	configLock.Lock()
	defer configLock.Unlock()

	if ConfigInstance == nil {
		ConfigInstance = NewConfig("test")

		// multiple locations to store the config file are supported by default
		// iterate over each one until a config is found. If by the end of the
		// loop no config is found in any of the locations, panic
		configFound := false
		for i := range commonConfigPaths {
			err := JSONToETLConfig(ConfigInstance, commonConfigPaths[i])
			if err == nil {
				ConfigInstance.Path = commonConfigPaths[i] // the path we found the config for future reference
				configFound = true
				break
			} else {
				fmt.Println("not found")
			}
		}

		// no config found
		if !configFound {
			panic("(!) the etl configuration file can either not be found or is corrupted")
		}
	}

	return ConfigInstance
}

func (c *Config) Store() bool {
	// verify that the config file we initially loaded from has not been deleted
	if _, err := os.Stat(c.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}

	jsonRepOfConfig, err := json.Marshal(c)
	if err != nil {
		return false
	}

	err = os.WriteFile(c.Path, jsonRepOfConfig, 0666)
	if err != nil {
		return false
	}
	return true
}

func NewCore() *Core {
	core := new(Core)

	core.C1 = make(chan DatabaseRequest)
	core.C2 = make(chan DatabaseResponse)
	core.C3 = make(chan MessengerRequest)
	core.C4 = make(chan MessengerResponse)
	core.C5 = make(chan ProvisionerRequest)
	core.C6 = make(chan ProvisionerResponse)
	core.C7 = make(chan DatabaseRequest)
	core.C8 = make(chan DatabaseResponse)
	core.C9 = make(chan CacheRequest)
	core.C10 = make(chan CacheResponse)
	core.interrupt = make(chan InterruptEvent)

	var ok bool
	core.HttpThread, ok = NewHttp(core.interrupt, core.C1, core.C2, core.C5, core.C6)
	if !ok {
		return nil
	}
	core.ProvisionerThread, ok = NewProvisioner(core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C9, core.C10)
	if !ok {
		return nil
	}
	core.MessengerThread, ok = NewMessenger(core.interrupt, core.C3, core.C4)
	if !ok {
		return nil
	}
	core.DatabaseThread, ok = NewDatabase(core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C7, core.C8)
	if !ok {
		return nil
	}
	core.CacheThread, ok = NewCacheThread(core.interrupt, core.C9, core.C10)
	if !ok {
		return nil
	}

	return core
}

func (core *Core) Run() {
	// needed in-case the proceeding core need logging or email capabilities during startup
	core.MessengerThread.Setup()
	go core.MessengerThread.Start() // event loop
	if GetConfigInstance().Debug {
		log.Println("(+) Messenger Thread Started")
	}

	// needed in-case the supervisor or http core need to populate Data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if GetConfigInstance().Debug {
		log.Println("(+) Database Thread Started")
	}

	// we need a way to provision clusters if we are receiving core before we can
	core.ProvisionerThread.Setup()
	go core.ProvisionerThread.Start() // event loop
	if GetConfigInstance().Debug {
		log.Println("(+) Provisioner Thread Started")
	}

	// if we chain requests, we should have a way to save that Data for re-use
	core.CacheThread.Setup()
	go core.CacheThread.Start()
	if GetConfigInstance().Debug {
		log.Println("(+) Cache Thread Started")
	}

	// the gateway to the frontend cluster should be the last startup
	core.HttpThread.Setup()
	go core.HttpThread.Start() // event loop
	if GetConfigInstance().Debug {
		log.Println("(+) RPC Thread Started")
	}

	// output all the static mounts on the system
	config := GetConfigInstance()
	numOfMountedClusters := len(config.AutoMount)

	output := "(!) Statically Mounted Cluster"
	if numOfMountedClusters > 1 {
		output += "s"
	}
	output += " [ "

	for idx, cluster := range config.AutoMount {
		output += cluster
		if idx == numOfMountedClusters {
			output += ", "
		}
	}
	output += " ]"

	// only output the statically mounted clusters if the debug tag is enabled
	if GetConfigInstance().Debug {
		log.Println(output)
	}

	// monitor system calls being sent to the process, if the etl is being
	// run on a local machine, the developer might attempt to kill the process with SIGINT
	// requiring us to cleanly close the application without risking the loss of Data
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block until we receive an interrupt from the system
		core.interrupt <- Panic
	}()

	// an interrupt can be sent by any thread that has access to the channel if an
	// error or end-state has been reached by the application
	switch <-core.interrupt {
	case Panic:
		log.Println("(IO) encountered panic")
		break
	default: // shutdown
		log.Println("(IO) shutting down")
		break
	}

	// close the gateway, stop new core from flooding into the servers
	core.HttpThread.Teardown()

	if GetConfigInstance().Debug {
		log.Println("(-) http shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	core.ProvisionerThread.Teardown()

	if GetConfigInstance().Debug {
		log.Println("(-) provisioner shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the Data is useless, shutdown
	core.CacheThread.Teardown()

	if GetConfigInstance().Debug {
		log.Println("(-) cache shutdown")
	}

	// the supervisor might need to store Data while finishing, close after
	core.DatabaseThread.Teardown()

	if GetConfigInstance().Debug {
		log.Println("(-) database shutdown")
	}

	// the preceding core might need to log, or send alerts of failure during shutdown
	core.MessengerThread.Teardown()

	if GetConfigInstance().Debug {
		log.Println("(-) messenger shutdown")
	}
}

func (core *Core) Cluster(identifier string, cluster cluster.Cluster, config ...cluster.Config) {
	p := GetProvisionerInstance()
	if len(config) > 0 {
		p.Register(identifier, cluster, config[0])
	} else {
		p.Register(identifier, cluster)
	}
}
