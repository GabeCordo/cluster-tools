package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/Sentmint/cluster-tools/internal/core"
	"log"
	"os"
)

type StartCommand struct {
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// check to see that the etl thread has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("missing configurations folder at %s\nmake sure you run 'ctgate init'\n", DefaultConfigsFolder)
		return commandline.Terminate
	}

	c, err := core.New(DefaultConfigFile)
	if err != nil {
		log.Panic(err.Error())
	}

	c.Run()

	return commandline.Terminate
}
