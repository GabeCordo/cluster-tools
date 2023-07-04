package main

import (
	"flag"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl/common"
	"github.com/GabeCordo/etl/core"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

var help = flag.Bool("help", false, "show help")
var includeCommonModuleFlag = false

func main() {

	modulePathFlag := os.Getenv("ETL_ENGINE_MODULES")

	if modulePathFlag == "" {
		panic("missing $ETL_ENGINE_MODULES variable in the environment")
	}

	configPathFlag := os.Getenv("ETL_ENGINE_CONFIG")

	if configPathFlag == "" {
		panic("missing $ETL_ENGINE_CONFIG variable in the environment")
	}

	flag.BoolVar(&includeCommonModuleFlag, "common", false, "load the common module in")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if configPathFlag == "" {
		log.Panic("you must specify a path to an etl confie file with '-config' or '--config'")
	}

	c, err := core.NewCore(configPathFlag)
	if err != nil {
		log.Panic(err.Error())
	}

	// load in the example cluster into the "common" module
	// ~ this may be helpful for people trying to spin up the framework for the first time and
	//   want to use this as an example of how to use it as an operator rather than a developer
	if includeCommonModuleFlag {
		Vec := common.Vector{}

		config := cluster.DefaultConfig
		config.Identifier = "Vec"
		c.Cluster("Vec", Vec, config)

		VecWait := common.VectorWait{}

		configWait := cluster.DefaultConfig
		configWait.Identifier = "VecWait"
		configWait.OnLoad = cluster.WaitAndPush
		c.Cluster("VecWait", VecWait, configWait)

		KeyTest := common.MetaDataCluster{}

		configMDC := cluster.DefaultConfig
		configMDC.Identifier = "KeyTest"
		c.Cluster("KeyTest", KeyTest, configMDC)
	}

	go func() {

		// wait for the framework to be brought up
		time.Sleep(1 * time.Second)

		// load in any pre-compiled modules before startup
		// ~ this allows us to 'statically' load them into the framework instance before it is
		//	 operational, also, avoiding the need to dynamically load them over HTTP one by one
		if modulePathFlag != "" {
			err := filepath.Walk(modulePathFlag, func(path string, info fs.FileInfo, err error) error {
				if info == nil {
					return nil
				}
				// the root folder will be included in the walk of the directory, we know this is not a module,
				// so we should skip the path if it is pointing to the root
				if info.IsDir() && (path != modulePathFlag) {
					c.Module(path)
				}
				return nil
			})
			if err != nil {
				log.Println("issue loading module")
			}
		}
	}()

	c.Run()
}
