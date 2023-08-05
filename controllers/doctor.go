package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl-light/core/config"
	"github.com/GabeCordo/etl/framework/core"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type DoctorCommand struct {
}

func (dc DoctorCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(core.DefaultFrameworkFolder); err != nil {
		fmt.Println("[x] etl has never been initialized, run 'etl init'")
		return commandline.Terminate
	}

	if _, err := os.Stat(core.DefaultModulesFolder); err != nil {
		fmt.Printf("[x] the modules folder is missing (%s)\n", core.DefaultModulesFolder)
	} else {
		fmt.Printf("[✓] the modules folder exists (%s)\n", core.DefaultModulesFolder)
	}

	if _, err := os.Stat(core.DefaultLogsFolder); err != nil {
		fmt.Printf("[x] the logs folder is missing (%s)\n", core.DefaultLogsFolder)
	} else {
		fmt.Printf("[✓] the logs folder exists (%s)\n", core.DefaultLogsFolder)
	}

	if _, err := os.Stat(core.DefaultConfigsFolder); err != nil {
		fmt.Printf("[x] the configs folder is missing (%s)\n", core.DefaultConfigsFolder)
	} else {
		fmt.Printf("[✓] the configs folder exists (%s)\n", core.DefaultConfigsFolder)
	}

	if _, err := os.Stat(core.DefaultConfigFile); err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", core.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common file exists (%s)\n", core.DefaultConfigFile)
	}

	configFile, err := os.Open(core.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common file is missing (%s)\n", core.DefaultConfigFile)
		return commandline.Terminate
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", core.DefaultConfigFile)
		return commandline.Terminate
	}

	c := &config.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", core.DefaultConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] the global common is healthy (%s)\n", core.DefaultConfigFile)
	}

	return commandline.Terminate
}
