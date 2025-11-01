package controllers

import (
	"fmt"
	"github.com/FortifiedCode/commandline"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
	"github.com/FortifiedCode/flock/internal/targets/core"
	"log"
	"os"
)

type StartCommand struct {
}

func (sc StartCommand) banner() {
	fmt.Println("   __ _            _    \n  / _| | ___   ___| | __\n | |_| |/ _ \\ / __| |/ /\n |  _| | (_) | (__|   < \n |_| |_|\\___/ \\___|_|\\_\\")
	fmt.Println("[+] " + terminal.Purple + "The Distributed Service Framework " + terminal.Reset)
	fmt.Println("[+]" + terminal.Purple + " by Gabriel Cordovado 2022-25" + terminal.Reset)
	fmt.Println()
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// check to see that the etl thread has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("missing configurations folder at %s\nmake sure you run 'flock init'\n", DefaultConfigsFolder)
		return commandline.Terminate
	}

	sc.banner()

	terminal.HorizontalBar()
	fmt.Println("Setting up dependencies...")

	c, err := core.New(DefaultCoreConfigFile)
	if err != nil {
		log.Panic(err.Error())
	}

	terminal.HorizontalBar()
	fmt.Println("Launching the threads...")

	c.Run()

	return commandline.Terminate
}
