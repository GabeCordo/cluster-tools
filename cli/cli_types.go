package cli

import "ETLFramework/core"

type CommandLine struct {
	Core  *core.Core
	Flags struct {
		Debug bool
	}
}
