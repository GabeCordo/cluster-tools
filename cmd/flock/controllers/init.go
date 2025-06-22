package controllers

import (
	"fmt"
	"os"

	"github.com/GabeCordo/Flock/internal/core"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
)

var (
	userCacheDir, _         = os.UserCacheDir()
	DefaultFrameworkFolder  = userCacheDir + "/flock/"
	DefaultConfigsFolder    = DefaultFrameworkFolder + "configs/"
	DefaultCoreConfigFile   = DefaultFrameworkFolder + "core.yml"
	DefaultLogsFolder       = DefaultFrameworkFolder + "logs/"
	DefaultStatisticsFolder = DefaultFrameworkFolder + "statistics/"
	DefaultSchedulesFolder  = DefaultFrameworkFolder + "schedules/"
	DefaultMessengerFolder  = DefaultFrameworkFolder + "messenger/"
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
		fmt.Printf("[%s!%s]flock has already been initialized\n",
			terminal.Red, terminal.Reset)
		return commandline.Terminate
	}

	fmt.Println("flock has not been initialized")

	if err := os.Mkdir(DefaultFrameworkFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultFrameworkFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created default cache folder %s\n",
			terminal.Green, terminal.Reset, DefaultFrameworkFolder)
	}

	if err := os.Mkdir(DefaultLogsFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultLogsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created logs folder %s\n",
			terminal.Green, terminal.Reset, DefaultLogsFolder)
	}

	if err := os.Mkdir(DefaultStatisticsFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultStatisticsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created statistics folder %s\n",
			terminal.Green, terminal.Reset, DefaultStatisticsFolder)
	}

	if err := os.Mkdir(DefaultSchedulesFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultSchedulesFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created scheduels folder %s\n",
			terminal.Green, terminal.Reset, DefaultSchedulesFolder)
	}

	if err := os.Mkdir(DefaultMessengerFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultMessengerFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created scheduels folder %s\n",
			terminal.Green, terminal.Reset, DefaultMessengerFolder)
	}

	if err := os.Mkdir(DefaultConfigsFolder, 0700); err != nil {
		fmt.Printf("[%sx%s] failed to create %s directory %s\n",
			terminal.Red, terminal.Reset, DefaultConfigsFolder, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created configs folder %s\n",
			terminal.Green, terminal.Reset, DefaultConfigsFolder)
	}

	dst, err := os.Create(DefaultCoreConfigFile) // #nosec G304 -- Constant is not user controlled
	if err != nil {
		fmt.Printf("[%sx%s] failed to create %s %s\n",
			terminal.Red, terminal.Reset, DefaultCoreConfigFile, err.Error())
		return commandline.Terminate
	}

	bytes, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Printf("[%sx%s] failed to marshal default common %s\n",
			terminal.Red, terminal.Reset, err.Error())
		return commandline.Terminate
	}

	if _, err := dst.Write(bytes); err != nil {
		fmt.Printf("[%sx%s] failed to write bytes of default common to file %s\n",
			terminal.Red, terminal.Reset, err.Error())
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] created default common %s\n",
			terminal.Green, terminal.Red, DefaultCoreConfigFile)
	}

	return commandline.Terminate
}
