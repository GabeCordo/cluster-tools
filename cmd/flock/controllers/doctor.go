package controllers

import (
	"fmt"
	"io"
	"os"

	"github.com/GabeCordo/Flock/internal/core"
	"github.com/GabeCordo/Flock/internal/terminal"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
)

type DoctorCommand struct {
}

func (dc DoctorCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(DefaultFrameworkFolder); err != nil {
		fmt.Printf("[%sx%s] flock has never been initialized, statistic 'flock init'\n",
			terminal.Red, terminal.Reset)
		return commandline.Terminate
	}

	if _, err := os.Stat(DefaultLogsFolder); err != nil {
		fmt.Printf("[%sx%s] the logs folder is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultLogsFolder)
	} else {
		fmt.Printf("[%s✓%s] the logs folder exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultLogsFolder)
	}

	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("[%sx%s] the configs folder is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultConfigsFolder)
	} else {
		fmt.Printf("[%s✓%s] the configs folder exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultConfigsFolder)
	}

	if _, err := os.Stat(DefaultStatisticsFolder); err != nil {
		fmt.Printf("[%sx%s] the statistics folder is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultStatisticsFolder)
	} else {
		fmt.Printf("[%s✓%s] the statistics folder exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultStatisticsFolder)
	}

	if _, err := os.Stat(DefaultSchedulesFolder); err != nil {
		fmt.Printf("[%sx%s] the scheduels folder is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultSchedulesFolder)
	} else {
		fmt.Printf("[%s✓%s] the scheduels folder exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultSchedulesFolder)
	}

	if _, err := os.Stat(DefaultMessengerFolder); err != nil {
		fmt.Printf("[%sx%s] the messenger folder is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultMessengerFolder)
	} else {
		fmt.Printf("[%s✓%s] the messenger folder exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultMessengerFolder)
	}

	if _, err := os.Stat(DefaultCoreConfigFile); err != nil {
		fmt.Printf("[%sx%s] the global common file is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultCoreConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] the global common file exists (%s)\n",
			terminal.Green, terminal.Reset, DefaultCoreConfigFile)
	}

	configFile, err := os.Open(DefaultCoreConfigFile)
	if err != nil {
		fmt.Printf("[%sx%s] the global common file is missing (%s)\n",
			terminal.Red, terminal.Reset, DefaultCoreConfigFile)
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
		fmt.Printf("[%sx%s] the global common is corrupt (%s)\n",
			terminal.Red, terminal.Reset, DefaultCoreConfigFile)
		return commandline.Terminate
	}

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[%sx%s] the global common is corrupt (%s)\n",
			terminal.Red, terminal.Reset, DefaultCoreConfigFile)
		return commandline.Terminate
	} else {
		fmt.Printf("[%s✓%s] the global common is healthy (%s)\n",
			terminal.Green, terminal.Reset, DefaultCoreConfigFile)
	}

	return commandline.Terminate
}
