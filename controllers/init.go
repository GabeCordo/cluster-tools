package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl-light/core"
	"github.com/GabeCordo/etl/core/threads"
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
	defaultConfig.Messenger.LogFiles.Directory = "/var/etl/logs"
	defaultConfig.Messenger.EnableSmtp = false
	defaultConfig.Net.Client.Host = "localhost"
	defaultConfig.Net.Client.Port = 8136
	defaultConfig.Net.Processor.Host = "localhost"
	defaultConfig.Net.Processor.Port = 8137
	defaultConfig.Path = threads.DefaultFrameworkFolder

	defaultConfig.Messenger.LogFiles.Directory = threads.DefaultLogsFolder

	if _, err := os.Stat(threads.DefaultFrameworkFolder); err == nil {
		fmt.Println("etl has already been initialized")
		return commandline.Terminate
	}

	fmt.Println("etl has not been initialized")

	if err := os.Mkdir(threads.DefaultFrameworkFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", threads.DefaultFrameworkFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created default cache folder %s\n", threads.DefaultFrameworkFolder)
	}

	if err := os.Mkdir(threads.DefaultModulesFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", threads.DefaultModulesFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created modules folder %s\n", threads.DefaultModulesFolder)
	}

	if err := os.Mkdir(threads.DefaultLogsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", threads.DefaultLogsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created logs folder %s\n", threads.DefaultLogsFolder)
	}

	if err := os.Mkdir(threads.DefaultStatisticsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", threads.DefaultStatisticsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", threads.DefaultStatisticsFolder)
	}

	if err := os.Mkdir(threads.DefaultConfigsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", threads.DefaultConfigsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", threads.DefaultConfigsFolder)
	}

	dst, err := os.Create(threads.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] failed to create %s %s\n", threads.DefaultConfigFile, err.Error())
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
		fmt.Printf("[✓] created default common %s\n", threads.DefaultConfigFile)
	}

	return commandline.Terminate
}
