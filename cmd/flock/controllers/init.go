package controllers

import (
	"fmt"
	"os"

	"github.com/GabeCordo/Flock/internal/core"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
)

var (
	userCacheDir, _            = os.UserCacheDir()
	DefaultFrameworkFolder     = userCacheDir + "/flock/"
	DefaultConfigsFolder       = DefaultFrameworkFolder + "configs/"
	DefaultCoreConfigFile      = DefaultFrameworkFolder + "core.yml"
	DefaultProcessorConfigFile = DefaultFrameworkFolder + "processor.yml"
	DefaultLogsFolder          = DefaultFrameworkFolder + "logs/"
	DefaultStatisticsFolder    = DefaultFrameworkFolder + "statistics/"
	DefaultSchedulesFolder     = DefaultFrameworkFolder + "schedules/"
	DefaultMessengerFolder     = DefaultFrameworkFolder + "messenger/"
)

type InitCommand struct {
}

func (ic InitCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	defaultConfig := core.Config{Debug: true, HardTerminateTime: 2}
	defaultConfig.EnableCors = false
	defaultConfig.EnableRepl = false
	defaultConfig.MountByDefault = true
	defaultConfig.MaxWaitForResponse = 2

	defaultConfig.Cache.Expiry = 2
	defaultConfig.Cache.MaxSize = 1000

	defaultConfig.Messenger.EnableLogging = true
	defaultConfig.Messenger.LogFiles.Directory = DefaultLogsFolder
	defaultConfig.Messenger.EnableSmtp = false

	defaultConfig.Net.Client.Host = "0.0.0.0"
	defaultConfig.Net.Client.Port = 8136
	defaultConfig.Net.Processor.Host = "0.0.0.0"
	defaultConfig.Net.Processor.Port = 8137

	defaultConfig.Paths.Root = DefaultFrameworkFolder
	defaultConfig.Paths.Configs = DefaultConfigsFolder
	defaultConfig.Paths.Statistics = DefaultStatisticsFolder
	defaultConfig.Paths.Schedules = DefaultSchedulesFolder
	defaultConfig.Paths.Messenger = DefaultMessengerFolder
	defaultConfig.Paths.Logs = DefaultLogsFolder

	defaultConfig.Processor.ProbeEvery = 1
	defaultConfig.Processor.MaxRetry = 10

	if _, err := os.Stat(DefaultFrameworkFolder); err == nil {
		fmt.Println("flock has already been initialized")
		return commandline.Terminate
	}

	fmt.Println("flock has not been initialized")

	if err := os.Mkdir(DefaultFrameworkFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultFrameworkFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created default cache folder %s\n", DefaultFrameworkFolder)
	}

	if err := os.Mkdir(DefaultLogsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultLogsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created logs folder %s\n", DefaultLogsFolder)
	}

	if err := os.Mkdir(DefaultStatisticsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultStatisticsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created statistics folder %s\n", DefaultStatisticsFolder)
	}

	if err := os.Mkdir(DefaultSchedulesFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultSchedulesFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created scheduels folder %s\n", DefaultSchedulesFolder)
	}

	if err := os.Mkdir(DefaultMessengerFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultMessengerFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created scheduels folder %s\n", DefaultMessengerFolder)
	}

	if err := os.Mkdir(DefaultConfigsFolder, 0700); err != nil {
		fmt.Printf("[x] failed to create %s directory %s\n", DefaultConfigsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[✓] created configs folder %s\n", DefaultConfigsFolder)
	}

	dst, err := os.Create(DefaultCoreConfigFile)
	if err != nil {
		fmt.Printf("[x] failed to create %s %s\n", DefaultCoreConfigFile, err.Error())
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
		fmt.Printf("[✓] created default common %s\n", DefaultCoreConfigFile)
	}

	return commandline.Terminate
}
