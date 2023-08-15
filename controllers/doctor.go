package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl-light/core/config"
	"github.com/GabeCordo/etl/core/threads"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type DoctorCommand struct {
}

func (dc DoctorCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(threads.DefaultFrameworkFolder); err != nil {
		fmt.Println("[x] etl has never been initialized, run 'etl init'")
		return commandline.Terminate
	}

	if _, err := os.Stat(threads.DefaultModulesFolder); err != nil {
		fmt.Printf("[x] the modules folder is missing (%s)\n", threads.DefaultModulesFolder)
	} else {
		fmt.Printf("[✓] the modules folder exists (%s)\n", threads.DefaultModulesFolder)
	}

	if _, err := os.Stat(threads.DefaultLogsFolder); err != nil {
		fmt.Printf("[x] the logs folder is missing (%s)\n", threads.DefaultLogsFolder)
	} else {
		fmt.Printf("[✓] the logs folder exists (%s)\n", threads.DefaultLogsFolder)
	}

	if _, err := os.Stat(threads.DefaultConfigsFolder); err != nil {
		fmt.Printf("[x] the configs folder is missing (%s)\n", threads.DefaultConfigsFolder)
	} else {
		fmt.Printf("[✓] the configs folder exists (%s)\n", threads.DefaultConfigsFolder)
	}

	if _, err := os.Stat(threads.DefaultStatisticsFolder); err != nil {
		fmt.Printf("[x] the statistics folder is missing (%s)\n", threads.DefaultStatisticsFolder)
	} else {
		fmt.Printf("[✓] the statistics folder exists (%s)\n", threads.DefaultStatisticsFolder)
	}

	if _, err := os.Stat(threads.DefaultConfigFile); err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", threads.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common file exists (%s)\n", threads.DefaultConfigFile)
	}

	configFile, err := os.Open(threads.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", threads.DefaultConfigFile)
		return commandline.Terminate
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", threads.DefaultConfigFile)
		return commandline.Terminate
	}

	c := &config.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", threads.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common is healthy (%s)\n", threads.DefaultConfigFile)
	}

	return commandline.Terminate
}
