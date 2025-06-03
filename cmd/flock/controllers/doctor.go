package controllers

import (
	"fmt"
	"io"
	"os"

	"github.com/GabeCordo/Flock/internal/core"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
)

type DoctorCommand struct {
}

func (dc DoctorCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(DefaultFrameworkFolder); err != nil {
		fmt.Println("[x] PipelineOps has never been initialized, statistic 'flock init'")
		return commandline.Terminate
	}

	if _, err := os.Stat(DefaultLogsFolder); err != nil {
		fmt.Printf("[x] the logs folder is missing (%s)\n", DefaultLogsFolder)
	} else {
		fmt.Printf("[✓] the logs folder exists (%s)\n", DefaultLogsFolder)
	}

	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("[x] the configs folder is missing (%s)\n", DefaultConfigsFolder)
	} else {
		fmt.Printf("[✓] the configs folder exists (%s)\n", DefaultConfigsFolder)
	}

	if _, err := os.Stat(DefaultStatisticsFolder); err != nil {
		fmt.Printf("[x] the statistics folder is missing (%s)\n", DefaultStatisticsFolder)
	} else {
		fmt.Printf("[✓] the statistics folder exists (%s)\n", DefaultStatisticsFolder)
	}

	if _, err := os.Stat(DefaultSchedulesFolder); err != nil {
		fmt.Printf("[x] the scheduels folder is missing (%s)\n", DefaultSchedulesFolder)
	} else {
		fmt.Printf("[✓] the scheduels folder exists (%s)\n", DefaultSchedulesFolder)
	}

	if _, err := os.Stat(DefaultMessengerFolder); err != nil {
		fmt.Printf("[x] the messenger folder is missing (%s)\n", DefaultMessengerFolder)
	} else {
		fmt.Printf("[✓] the messenger folder exists (%s)\n", DefaultMessengerFolder)
	}

	if _, err := os.Stat(DefaultConfigFile); err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common file exists (%s)\n", DefaultConfigFile)
	}

	configFile, err := os.Open(DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(configFile)

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	}

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common is healthy (%s)\n", DefaultConfigFile)
	}

	return commandline.Terminate
}
