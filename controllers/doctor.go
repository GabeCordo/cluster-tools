package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/mango/core"
	"github.com/GabeCordo/mango/core/threads/common"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type DoctorCommand struct {
}

func (dc DoctorCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(common.DefaultFrameworkFolder); err != nil {
		fmt.Println("[x] mango has never been initialized, run 'mango init'")
		return commandline.Terminate
	}

	if _, err := os.Stat(common.DefaultLogsFolder); err != nil {
		fmt.Printf("[x] the logs folder is missing (%s)\n", common.DefaultLogsFolder)
	} else {
		fmt.Printf("[✓] the logs folder exists (%s)\n", common.DefaultLogsFolder)
	}

	if _, err := os.Stat(common.DefaultConfigsFolder); err != nil {
		fmt.Printf("[x] the configs folder is missing (%s)\n", common.DefaultConfigsFolder)
	} else {
		fmt.Printf("[✓] the configs folder exists (%s)\n", common.DefaultConfigsFolder)
	}

	if _, err := os.Stat(common.DefaultStatisticsFolder); err != nil {
		fmt.Printf("[x] the statistics folder is missing (%s)\n", common.DefaultStatisticsFolder)
	} else {
		fmt.Printf("[✓] the statistics folder exists (%s)\n", common.DefaultStatisticsFolder)
	}

	if _, err := os.Stat(common.DefaultSchedulesFolder); err != nil {
		fmt.Printf("[x] the scheduels folder is missing (%s)\n", common.DefaultSchedulesFolder)
	} else {
		fmt.Printf("[✓] the scheduels folder exists (%s)\n", common.DefaultSchedulesFolder)
	}

	if _, err := os.Stat(common.DefaultConfigFile); err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common file exists (%s)\n", common.DefaultConfigFile)
	}

	configFile, err := os.Open(common.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	}

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common is healthy (%s)\n", common.DefaultConfigFile)
	}

	return commandline.Terminate
}
