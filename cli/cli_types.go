package cli

import (
	"ETLFramework/core"
	"fmt"
)

type Printables interface {
	String() string
}

type CommandLine struct {
	Core  *core.Core
	Flags struct {
		Debug bool
	}
}

type Vector struct {
	x int
	y int
}

func (v Vector) String() string {
	return fmt.Sprintf("(%d,%d)", v.x, v.y)
}
