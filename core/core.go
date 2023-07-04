package core

import (
	"bufio"
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/config"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/components/provisioner"
	"github.com/GabeCordo/etl/components/utils"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

const (
	Version string = "v0.1.9-alpha"
)

var (
	configLock     = &sync.Mutex{}
	ConfigInstance *config.Config
)

func GetDefaultConfigPath() string {

	if runtime.GOOS == "windows" {
		return "%PROGRAMDATA%/etl/config.etl.yaml"
	} else if runtime.GOOS == "linux" {
		return "/opt/etl/config.etl.yaml"
	} else {
		return "/etc/etl/config.etl.yaml"
	}
}

func GetConfigInstance(configPath ...string) *config.Config {
	configLock.Lock()
	defer configLock.Unlock()

	/* if this is the first time the config is being loaded the develoepr
	   needs to pass in a configPath to load the config instance from
	*/
	if (ConfigInstance == nil) && (len(configPath) < 1) {
		return nil
	}

	if ConfigInstance == nil {
		ConfigInstance = config.NewConfig("test")

		if err := config.YAMLToETLConfig(ConfigInstance, configPath[0]); err == nil {
			// the configPath we found the config for future reference
			ConfigInstance.Path = configPath[0]
			// if the MaxWaitForResponse is not set, then simply default to 2.0
			if ConfigInstance.MaxWaitForResponse == 0 {
				ConfigInstance.MaxWaitForResponse = 2
			}
		} else {
			log.Println("(!) the etl configuration file can either not be found or is corrupted")
			log.Fatal(fmt.Sprintf("%s was not a valid config path\n", configPath))
		}
	}

	return ConfigInstance
}

func NewCore(configPath string) (*Core, error) {
	core := new(Core)

	core.C1 = make(chan threads.DatabaseRequest)
	core.C2 = make(chan threads.DatabaseResponse)
	core.C3 = make(chan threads.MessengerRequest)
	core.C4 = make(chan threads.MessengerResponse)
	core.C5 = make(chan threads.ProvisionerRequest)
	core.C6 = make(chan threads.ProvisionerResponse)
	core.C7 = make(chan threads.DatabaseRequest)
	core.C8 = make(chan threads.DatabaseResponse)
	core.C9 = make(chan threads.CacheRequest)
	core.C10 = make(chan threads.CacheResponse)
	core.C11 = make(chan threads.MessengerRequest)
	core.interrupt = make(chan threads.InterruptEvent)

	/* load the config in for the first time */
	if config := GetConfigInstance(configPath); config == nil {
		log.Panic("could not create config")
	}

	httpLogger, err := utils.NewLogger(utils.Http, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.HttpThread, err = NewHttp(httpLogger, core.interrupt, core.C1, core.C2, core.C5, core.C6)
	if err != nil {
		return nil, err
	}

	provisionerLogger, err := utils.NewLogger(utils.Provisioner, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.ProvisionerThread, err = NewProvisioner(provisionerLogger, core.interrupt, core.C5, core.C6, core.C7, core.C8, core.C9, core.C10, core.C11)
	if err != nil {
		return nil, err
	}

	messengerLogger, err := utils.NewLogger(utils.Messenger, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.MessengerThread, err = NewMessenger(messengerLogger, core.interrupt, core.C3, core.C4, core.C11)
	if err != nil {
		return nil, err
	}

	databaseLogger, err := utils.NewLogger(utils.Database, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.DatabaseThread, err = NewDatabase(databaseLogger, core.interrupt, core.C1, core.C2, core.C3, core.C4, core.C7, core.C8)
	if err != nil {
		return nil, err
	}

	cacheLogger, err := utils.NewLogger(utils.Database, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.CacheThread, err = NewCacheThread(cacheLogger, core.interrupt, core.C9, core.C10)
	if err != nil {
		return nil, err
	}

	coreLogger, err := utils.NewLogger(utils.Undefined, &GetConfigInstance().Debug)
	if err != nil {
		return nil, err
	}
	core.logger = coreLogger

	return core, nil
}

func (core *Core) Banner() {
	fmt.Println("   ___    _____    _")
	fmt.Println("  | __|  |_   _|  | |")
	fmt.Println("  | _|     | |    | |__")
	fmt.Println("  |___|   _|_|_   |____|")
	fmt.Println("_|\"\"\"\"\"|_|\"\"\"\"\"|_|\"\"\"\"\"|")
	fmt.Println("\"`-0-0-'\"`-0-0-'\"`-0-0-'")
	fmt.Println("[+] " + utils.Purple + "Extract Transform Load Framework " + utils.Reset + Version)
	fmt.Println("[+]" + utils.Purple + " by Gabriel Cordovado 2022-23" + utils.Reset)
	fmt.Println()
}

func (core *Core) Run() {

	core.Banner()

	core.logger.SetColour(utils.Purple)

	if GetConfigInstance().Debug {
		core.logger.Println("debug mode ON")
	} else {
		core.logger.Println("debug mode OFF")
	}

	// needed in-case the proceeding core need logging or email capabilities during startup
	core.MessengerThread.Setup()
	go core.MessengerThread.Start() // event loop
	if GetConfigInstance().Debug {
		core.logger.Println("Messenger Thread Started")
		//log.Println(utils.Purple + "(+)" + utils.Reset + " Messenger Thread Started")
	}

	// needed in-case the supervisor or http core need to populate Data on startup
	core.DatabaseThread.Setup()
	go core.DatabaseThread.Start() // event loop
	if GetConfigInstance().Debug {
		core.logger.Println("Database Thread Started")
		//log.Println(utils.Purple + "(+)" + utils.Reset + " Database Thread Started")
	}

	// we need a way to provision common if we are receiving core before we can
	core.ProvisionerThread.Setup()
	go core.ProvisionerThread.Start() // event loop
	if GetConfigInstance().Debug {
		core.logger.Println("Provisioner Thread Started")
		//log.Println(utils.Purple + "(+)" + utils.Reset + " Provisioner Thread Started")
	}

	// if we chain requests, we should have a way to save that Data for re-use
	core.CacheThread.Setup()
	go core.CacheThread.Start()
	if GetConfigInstance().Debug {
		core.logger.Println("Cache Thread Started")
		//log.Println(utils.Purple + "(+)" + utils.Reset + " Cache Thread Started")
	}

	// the gateway to the frontend cluster should be the last startup
	core.HttpThread.Setup()
	go core.HttpThread.Start() // event loop
	if GetConfigInstance().Debug {
		core.logger.Println("HTTP API Thread Started")
		core.logger.Printf("\t- Listening on %s:%d\n", GetConfigInstance().Net.Host, GetConfigInstance().Net.Port)
		//log.Println(utils.Purple + "(+)" + utils.Reset + " HTTP API Thread Started")
		//log.Printf("\t- Listening on %s:%d\n", GetConfigInstance().Net.Host, GetConfigInstance().Net.Port)
	}

	// monitor system calls being sent to the process, if the etl is being
	// run on a local machine, the developer might attempt to kill the process with SIGINT
	// requiring us to cleanly close the application without risking the loss of Data
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block until we receive an interrupt from the system
		core.interrupt <- threads.Panic
	}()

	go func() {
		fmt.Println()
		fmt.Println("the interactive shell is an experimental feature that is still being worked on. " +
			"there may be some issues or missing features that are under development.")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("@etl ")
			text, _ := reader.ReadString('\n')
			text = strings.ReplaceAll(text, "\n", "")

			if text == "modules" {
				p := GetProvisionerInstance()
				modules := p.GetModules()

				for _, module := range modules {
					module.Print()
				}
			}
		}
	}()

	// an interrupt can be sent by any thread that has access to the channel if an
	// error or end-state has been reached by the application
	switch <-core.interrupt {
	case threads.Panic:
		core.logger.Printf("[IO] %s\n", " encountered panic")
		//log.Println(utils.Red + "(IO)" + utils.Reset + " encountered panic")
		break
	default: // shutdown
		core.logger.Printf("[IO] %s\n", " shutting down")
		//log.Println(utils.Red + "(IO)" + utils.Reset + " shutting down")
		break
	}

	// close the gateway, stop new core from flooding into the servers
	core.HttpThread.Teardown()

	core.logger.SetColour(utils.Red)

	if GetConfigInstance().Debug {
		core.logger.Println("http shutdown")
		//log.Println(utils.Red + "(-)" + utils.Reset + " http shutdown")
	}

	// THIS WILL TAKE THE LONGEST - clean channels and finish processing
	core.ProvisionerThread.Teardown()

	if GetConfigInstance().Debug {
		core.logger.Println("provisioner shutdown")
		//log.Println(utils.Red + "(-)" + utils.Reset + " provisioner shutdown")
	}

	// we won't need the cache if the cluster thread is shutdown, the Data is useless, shutdown
	core.CacheThread.Teardown()

	if GetConfigInstance().Debug {
		core.logger.Println("cache shutdown")
		//log.Println(utils.Red + "(-)" + utils.Reset + " cache shutdown")
	}

	// the supervisor might need to store Data while finishing, close after
	core.DatabaseThread.Teardown()

	if GetConfigInstance().Debug {
		core.logger.Println("database shutdown")
		//log.Println(utils.Red + "(-)" + utils.Reset + " database shutdown")
	}

	// the preceding core might need to log, or send alerts of failure during shutdown
	core.MessengerThread.Teardown()

	if GetConfigInstance().Debug {
		core.logger.Println("messenger shutdown")
		//log.Println(utils.Red + "(-)" + utils.Reset + " messenger shutdown")
	}
}

func (core *Core) Cluster(identifier string, implementation cluster.Cluster, config ...cluster.Config) {

	p := GetProvisionerInstance()
	defaultModule, _ := p.GetModule(provisioner.DefaultFrameworkModule) // the default framework module should always be found

	clusterWrapper, _ := defaultModule.AddCluster(identifier, implementation)
	clusterWrapper.Mount()

	if len(config) == 1 {
		d := GetDatabaseInstance()
		d.StoreClusterConfig(provisioner.DefaultFrameworkModule, config[0])
	}
}

func (core *Core) Module(path string) (success bool, description string) {

	success, description = RegisterModule(core.HttpThread.C5, core.HttpThread.provisionerResponseTable, path)
	return success, description
}
