package client

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/core"
)

// MOUNT COMMAND

type MountCommand struct {
	PublicName string
}

func (mc MountCommand) Name() string {
	return mc.PublicName
}

func (mc MountCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	if cl.Flags.Show {
		// show all the clusters mounted to the config (by default)
		mc.ShowMountClusters(cl)
	} else if cl.Flags.Create {
		// add a new mount to the config, so that it IS mounted on startup
		mc.AddMountCluster(cl)
	} else if cl.Flags.Delete {
		// remove a mount from the config, so it is NOT mounted on startup
		mc.RemoveMountCluster(cl)
	}

	return true // complete
}

func (mc MountCommand) AddMountCluster(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	projectPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory)
	etlConfigFile := projectPath.File("config.etl.json")

	if etlConfigFile.DoesNotExist() {
		fmt.Println("(!) there is not ETL project in this directory")
		return true
	}

	clusterName := cl.NextArg()
	if clusterName == commandline.FinalArg {
		fmt.Println("(!) a cluster PublicName was not provided")
		return true
	}

	config := core.Config{}
	core.JSONToETLConfig(&config, etlConfigFile.ToString())

	config.AutoMount = append(config.AutoMount, clusterName)

	config.ToJson(etlConfigFile.ToString())

	return true // complete
}

func (mc MountCommand) RemoveMountCluster(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	projectPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory)
	etlConfigFile := projectPath.File("config.etl.json")

	if etlConfigFile.DoesNotExist() {
		fmt.Println("there is not ETL project in this directory")
		return true
	}

	clusterName := cl.NextArg()
	if clusterName == commandline.FinalArg {
		fmt.Println("a cluster PublicName was not provided")
		return true
	}

	config := core.Config{}
	core.JSONToETLConfig(&config, etlConfigFile.ToString())

	clusterMountFound := false
	modifiedClusterAutoMountList := []string{}
	for _, cluster := range config.AutoMount {
		if cluster == clusterName {
			clusterMountFound = true
		} else {
			modifiedClusterAutoMountList = append(modifiedClusterAutoMountList, cluster)
		}
	}

	if !clusterMountFound {
		fmt.Println("no cluster was mounted with that PublicName")
		return true
	}

	config.AutoMount = modifiedClusterAutoMountList // assign the new slice without the cluster

	config.ToJson(etlConfigFile.ToString())

	return true // complete
}

func (mc MountCommand) ShowMountClusters(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	projectPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory)
	etlConfigFile := projectPath.File("config.etl.json")

	if etlConfigFile.DoesNotExist() {
		fmt.Println("there is not ETL project in this directory")
		return true
	}

	config := core.Config{}
	core.JSONToETLConfig(&config, etlConfigFile.ToString())

	if len(config.AutoMount) == 0 {
		fmt.Println("no clusters are auto mounted")
		return true
	}

	fmt.Println("the following clusters are auto mounted to the project:")
	for _, cluster := range config.AutoMount {
		fmt.Println(cluster)
	}

	return true // complete
}
