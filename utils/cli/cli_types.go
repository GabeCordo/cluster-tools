package cli

import (
	"etl/core"
	"os"
	"time"
)

type Config struct {
	Version  float32             `json:"version"`
	Projects map[string]struct { // let the key be the host identifier
		name      string `json:"name"`
		host      string `json:"host"`
		port      int    `json:"port"`
		publicKey string `json:"key"`
	} `json:"projects"`
	UserProfile struct {
		FirstName string `json:"first-name"`
		LastName  string `json:"last-name"`
		Email     string `json:"email"`
	} `json:"profile"`
}

type Terminate bool

type Command interface {
	Name() string
	Run(cli *CommandLine) Terminate
}

const (
	FinalArg               string      = ""
	DefaultFilePermissions os.FileMode = 0755
)

type CommandLine struct {
	Core *core.Core

	Config *Config

	Flags struct {
		Debug  bool
		Create bool
		Delete bool
		Show   bool
	}

	MetaData struct {
		WorkingDirectory string
		TimeCalled       time.Time
	}

	args        []string
	numOfArgs   int
	argsPointer int

	Commands map[string]Command
}
