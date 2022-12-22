package client

import (
	"fmt"
	"github.com/GabeCordo/commandline"
)

// VERSION COMMAND START

type VersionCommand struct {
	PublicName string
}

func (vc VersionCommand) Name() string {
	return vc.PublicName
}

func (vc VersionCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	fmt.Println(Version(cl))
	return true
}
