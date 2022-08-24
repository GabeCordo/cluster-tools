package cli

import (
	"ETLFramework/core"
	"fmt"
	"os"
)

func NewCommandLine(core *core.Core) (*CommandLine, bool) {
	if core == nil {
		return nil, false
	}

	cli := new(CommandLine)
	cli.Core = core

	return cli, true
}

func (cli *CommandLine) Run() {
	// start reading cli arguments
	args := os.Args[1:] // strip out the file descriptor in position 0
	for i := range args {
		if args[i] == "-h" || args[i] == "--help" {
			HelpCommand()
			return
		} else if args[i] == "-d" || args[i] == "--debug" {
			cli.Flags.Debug = true
		} else if args[i] == "-g" || args[i] == "--generate-key" {
			GenerateKeyPair()
			return
		} else if args[i] == "-i" || args[i] == "--interactive" {
			InteractiveDashboard()
			return
		}
	}
	// stop reading cli arguments

	if cli.Flags.Debug {
		fmt.Println("starting up etlframework..")
	}

	cli.Core.Run()

	if cli.Flags.Debug {
		fmt.Println("shutting down etlframework")
	}
}
