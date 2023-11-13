package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/mango/core"
	"github.com/GabeCordo/mango/core/threads/common"
	"log"
	"os"
)

type StartCommand struct {
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// check to see that the etl threads has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(common.DefaultConfigsFolder); err != nil {
		fmt.Println("the mango threads has never been initialized, run 'etl init'")
		return commandline.Terminate
	}

	c, err := core.New(common.DefaultConfigFile)
	if err != nil {
		log.Panic(err.Error())
	}

	// TODO : move to the processor
	// load in the example cluster into the "clusters" module
	// ~ this may be helpful for people trying to spin up the threads for the first time and
	//   want to use this as an example of how to use it as an operator rather than a developer
	//Vec := clusters.VectorCluster{}
	//
	//config := cluster.DefaultConfig
	//config.Identifier = "Vec"
	//c.Cluster("Vec", cluster.Batch, &Vec, config)
	//
	//VecWait := clusters.VectorWaitCluster{}
	//
	//configWait := cluster.DefaultConfig
	//configWait.Identifier = "VecWait"
	//configWait.OnLoad = cluster.WaitAndPush
	//c.Cluster("VecWait", cluster.Batch, &VecWait, configWait)
	//
	//KeyTest := clusters.MetaDataCluster{}
	//
	//configMDC := cluster.DefaultConfig
	//configMDC.Identifier = "KeyTest"
	//c.Cluster("KeyTest", cluster.Batch, &KeyTest, configMDC)

	// TODO : move to processor
	//go func() {
	//
	//	// wait for the threads to be brought up
	//	time.Sleep(1 * time.Second)
	//
	//	// load in any pre-compiled modules before startup
	//	// ~ this allows us to 'statically' load them into the threads instance before it is
	//	//	 operational, also, avoiding the need to dynamically load them over HTTP one by one
	//	err := filepath.Walk(threads.DefaultModulesFolder, func(path string, info fs.FileInfo, err error) error {
	//		if info == nil {
	//			return nil
	//		}
	//		// the root folder will be included in the walk of the directory, we know this is not a module,
	//		// so we should skip the path if it is pointing to the root
	//		if info.IsDir() && (path != threads.DefaultModulesFolder) {
	//			c.Module(path)
	//		}
	//		return nil
	//	})
	//	if err != nil {
	//		log.Println("issue loading module")
	//	}
	//}()

	c.Run()

	return commandline.Terminate
}
