package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/mango/core"
	"github.com/GabeCordo/mango/core/threads/common"
	"gopkg.in/yaml.v3"
	"os"
)

type InitCommand struct {
}

func (ic InitCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	defaultConfig := core.Config{Debug: true, HardTerminateTime: 2}
	defaultConfig.Cache.Expiry = 2
	defaultConfig.Cache.MaxSize = 1000
	defaultConfig.EnableCors = false
	defaultConfig.EnableRepl = false
	defaultConfig.Messenger.EnableLogging = true
	defaultConfig.Messenger.LogFiles.Directory = "/var/mangoose/logs"
	defaultConfig.Messenger.EnableSmtp = false
	defaultConfig.Net.Client.Host = "0.0.0.0"
	defaultConfig.Net.Client.Port = 8136
	defaultConfig.Net.Processor.Host = "0.0.0.0"
	defaultConfig.Net.Processor.Port = 8137
	defaultConfig.Path = common.DefaultFrameworkFolder

	defaultConfig.Messenger.LogFiles.Directory = common.DefaultLogsFolder

	if _, err := os.Stat(common.DefaultFrameworkFolder); err == nil {
		fmt.Println("mango has already been initialized")
		return commandline.Terminate
	}

	fmt.Println("mango has not been initialized")

	if err := os.Mkdir(common.DefaultFrameworkFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", common.DefaultFrameworkFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created default cache folder %s\n", common.DefaultFrameworkFolder)
	}

	if err := os.Mkdir(common.DefaultLogsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", common.DefaultLogsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created logs folder %s\n", common.DefaultLogsFolder)
	}

	if err := os.Mkdir(common.DefaultStatisticsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", common.DefaultStatisticsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created statistics folder %s\n", common.DefaultStatisticsFolder)
	}

	if err := os.Mkdir(common.DefaultSchedulesFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", common.DefaultSchedulesFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created scheduels folder %s\n", common.DefaultSchedulesFolder)
	}

	if err := os.Mkdir(common.DefaultConfigsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", common.DefaultConfigsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", common.DefaultConfigsFolder)
	}

	dst, err := os.Create(common.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] failed to create %s %s\n", common.DefaultConfigFile, err.Error())
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
		fmt.Printf("[✓] created default common %s\n", common.DefaultConfigFile)
	}

	return commandline.Terminate
}
