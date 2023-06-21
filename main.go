package main

import (
	"flag"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl/core"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

var help = flag.Bool("help", false, "show help")
var configPathFlag = ""
var modulePathFlag = ""
var vectorBoolFlag = false

func main() {

	flag.StringVar(&configPathFlag, "config", "", "declare a path to a config file")
	flag.StringVar(&modulePathFlag, "modules", "", "declare a path to a folder to laod modules statically")
	flag.BoolVar(&vectorBoolFlag, "vectors", false, "load in the vectors example cluster")

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
	if vectorBoolFlag {
		Vec := Vector{}

		config := cluster.DefaultConfig
		config.Identifier = "Vec"
		c.Cluster("Vec", Vec, config)
	}

	go func() {

		// wait for the framework to be brought up
		time.Sleep(1 * time.Second)

		// load in any pre-compiled modules before startup
		// ~ this allows us to 'statically' load them into the framework instance before it is
		//	 operational, also, avoiding the need to dynamically load them over HTTP one by one
		if modulePathFlag != "" {
			err := filepath.Walk(modulePathFlag, func(path string, info fs.FileInfo, err error) error {
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
