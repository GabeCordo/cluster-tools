package main

import (
	"etl/core"
	"etl/utils/cli"
)

func main() {

	c := core.NewCore()
	if commandLine, ok := cli.NewCommandLine(c); ok {
		commandLine.Run()
	}
}
