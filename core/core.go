package core

import (
	"ETLFramework/cluster"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	commonConfigPaths = [...]string{"config.etl.json", "/config/config.etl.json", "../config.etl.json", "../config/config.etl.json"}
	configLock        = &sync.Mutex{}
	ConfigInstance    *Config
)

func GetConfigInstance() *Config {
	configLock.Lock()
	defer configLock.Unlock()

	if ConfigInstance == nil {
		ConfigInstance = new(Config)

		// multiple locations to store the config file are supported by default
		// iterate over each one until a config is found. If by the end of the
		// loop no config is found in any of the locations, panic
		configFound := false
		for i := range commonConfigPaths {
			err := JSONToETLConfig(ConfigInstance, commonConfigPaths[i])
			if err == nil {
				configFound = true
				break
			}
		}

		// no config found
		if !configFound {
			panic("(!) missing etl configuration file")
		}
	}

	return ConfigInstance
}

func NewCore() *Core {
	core := new(Core)

	core.c1 = make(chan DatabaseRequest)
	core.c2 = make(chan DatabaseResponse)
	core.c3 = make(chan MessengerRequest)
	core.c4 = make(chan MessengerResponse)
	core.c5 = make(chan SupervisorRequest)
	core.c6 = make(chan SupervisorResponse)
	core.c7 = make(chan DatabaseRequest)
	core.c8 = make(chan DatabaseResponse)
	core.interrupt = make(chan InterruptEvent)

	var ok bool
	core.httpThread, ok = NewHttp(core.interrupt, core.c1, core.c2, core.c5, core.c6)
	if !ok {
		return nil
	}
	core.supervisorThread, ok = NewSupervisor(core.interrupt, core.c5, core.c6, core.c7, core.c8)
	if !ok {
		return nil
	}
	core.messengerThread, ok = NewMessenger(core.interrupt, core.c3, core.c4)
	if !ok {
		return nil
	}
	core.databaseThread, ok = NewDatabase(core.interrupt, core.c1, core.c2, core.c3, core.c4, core.c7, core.c8)
	if !ok {
		return nil
	}

	return core
}

func (core *Core) Run() {
	// needed in-case the proceeding core need logging or email capabilities during startup
	core.messengerThread.Setup()
	go core.messengerThread.Start() // event loop

	// needed in-case the supervisor or http core need to populate data on startup
	core.databaseThread.Setup()
	go core.databaseThread.Start() // event loop

	// we need a way to provision clusters if we are receiving core before we can
	core.supervisorThread.Setup()
	go core.supervisorThread.Start() // event loop

	// the gateway to the frontend cluster should be the last startup
	core.httpThread.Setup()
	go core.httpThread.Start() // event loop

	// monitor system calls being sent to the process, if the ETLFramework is being
	// run on a local machine, the developer might attempt to kill the process with SIGINT
	// requiring us to cleanly close the application without risking the loss of data
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
	core.httpThread.Teardown()

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	core.supervisorThread.Teardown()

	// the supervisor might need to store data while finishing, close after
	core.databaseThread.Teardown()

	// the preceding core might need to log, or send alerts of failure during shutdown
	core.messengerThread.Teardown()
}

func (core *Core) Cluster(identifier string, cluster cluster.Cluster, config ...cluster.Config) {
	s := GetSupervisorInstance()
	if len(config) > 0 {
		s.Register(identifier, cluster, config[0])
	} else {
		s.Register(identifier, cluster)
	}
}