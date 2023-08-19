package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/mango/core"
	"gopkg.in/yaml.v3"
	"os"
)

type InitCommand struct {
}

func (ic InitCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	defaultConfig := core.Config{Debug: true, HardTerminateTime: 2}
	defaultConfig.Cache.Expiry = 2
	defaultConfig.Cache.MaxSize = 1000
	defaultConfig.Messenger.EnableLogging = true
	defaultConfig.Messenger.LogFiles.Directory = "/var/mangoose/logs"
	defaultConfig.Messenger.EnableSmtp = false
	defaultConfig.Net.Client.Host = "localhost"
	defaultConfig.Net.Client.Port = 8136
	defaultConfig.Net.Processor.Host = "localhost"
	defaultConfig.Net.Processor.Port = 8137
	defaultConfig.Path = core.DefaultFrameworkFolder

	defaultConfig.Messenger.LogFiles.Directory = core.DefaultLogsFolder

	if _, err := os.Stat(core.DefaultFrameworkFolder); err == nil {
		fmt.Println("etl has already been initialized")
		return commandline.Terminate
	}

	fmt.Println("etl has not been initialized")

	if err := os.Mkdir(core.DefaultFrameworkFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", core.DefaultFrameworkFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created default cache folder %s\n", core.DefaultFrameworkFolder)
	}

	if err := os.Mkdir(core.DefaultLogsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", core.DefaultLogsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created logs folder %s\n", core.DefaultLogsFolder)
	}

	if err := os.Mkdir(core.DefaultStatisticsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", core.DefaultStatisticsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", core.DefaultStatisticsFolder)
	}

	if err := os.Mkdir(core.DefaultConfigsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", core.DefaultConfigsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", core.DefaultConfigsFolder)
	}

	dst, err := os.Create(core.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] failed to create %s %s\n", core.DefaultConfigFile, err.Error())
		return commandline.Terminate
	}

	bytes, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Printf("[x] failed to marshal default common %s\n", err.Error())
		return commandline.Terminate
	} else {

	}

	if _, err := dst.Write(bytes); err != nil {
		fmt.Printf("[x] failed to write bytes of default common to file %s\n", err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created default common %s\n", core.DefaultConfigFile)
	}

	return commandline.Terminate
}
