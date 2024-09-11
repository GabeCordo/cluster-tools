package core

import (
	"fmt"
	"github.com/GabeCordo/toolchain/logging"
)

func (core *Core) banner() {
	fmt.Println("        __           __            __              __    \n  _____/ /_  _______/ /____  _____/ /_____  ____  / /____\n / ___/ / / / / ___/ __/ _ \\/ ___/ __/ __ \\/ __ \\/ / ___/\n/ /__/ / /_/ (__  ) /_/  __/ /  / /_/ /_/ / /_/ / (__  ) \n\\___/_/\\__,_/____/\\__/\\___/_(_) \\__/\\____/\\____/_/____/ ")
	fmt.Println("[+] " + logging.Purple + "Cluster.tools Cloud Framework " + logging.Reset + Version)
	fmt.Println("[+]" + logging.Purple + " by Gabriel Cordovado 2022-24" + logging.Reset)
	fmt.Println()
}
