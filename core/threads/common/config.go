package common

import (
	"fmt"
	"github.com/GabeCordo/etl-light/core"
	"log"
	"runtime"
	"sync"
)

var (
	configLock     = &sync.Mutex{}
	ConfigInstance *core.Config
)

func GetDefaultConfigPath() string {

	if runtime.GOOS == "windows" {
		return "%PROGRAMDATA%/etl/common.etl.yaml"
	} else if runtime.GOOS == "linux" {
		return "/opt/etl/common.etl.yaml"
	} else {
		return "/etc/etl/common.etl.yaml"
	}
}

func GetConfigInstance(configPath ...string) *core.Config {
	configLock.Lock()
	defer configLock.Unlock()

	/* if this is the first time the common is being loaded the develoepr
	   needs to pass in a configPath to load the common instance from
	*/
	if (ConfigInstance == nil) && (len(configPath) < 1) {
		return nil
	}

	if ConfigInstance == nil {
		ConfigInstance = core.NewConfig("test")

		if err := core.YAMLToETLConfig(ConfigInstance, configPath[0]); err == nil {
			// the configPath we found the common for future reference
			ConfigInstance.Path = configPath[0]
			// if the MaxWaitForResponse is not set, then simply default to 2.0
			if ConfigInstance.MaxWaitForResponse == 0 {
				ConfigInstance.MaxWaitForResponse = 2
			}
		} else {
			log.Println("(!) the etl configuration file can either not be found or is corrupted")
			log.Fatal(fmt.Sprintf("%s was not a valid common path\n", configPath))
		}
	}

	return ConfigInstance
}
