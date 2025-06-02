package core

import (
	"fmt"

	"github.com/GabeCordo/toolchain/logging"
)

func (core *Core) banner() {
	//fmt.Println("    ____  _            ___            ____            \n   / __ \\(_)___  ___  / (_)___  ___  / __ \\____  _____\n  / /_/ / / __ \\/ _ \\/ / / __ \\/ _ \\/ / / / __ \\/ ___/\n / ____/ / /_/ /  __/ / / / / /  __/ /_/ / /_/ (__  ) \n/_/   /_/ .___/\\___/_/_/_/ /_/\\___/\\____/ .___/____/  \n       /_/                             /_/           ")
	fmt.Println("   __ _            _    \n  / _| | ___   ___| | __\n | |_| |/ _ \\ / __| |/ /\n |  _| | (_) | (__|   < \n |_| |_|\\___/ \\___|_|\\_\\")
	fmt.Println("[+] " + logging.Purple + "The Distributed Service Framework " + logging.Reset + Version)
	fmt.Println("[+]" + logging.Purple + " by Gabriel Cordovado 2022-25" + logging.Reset)
	fmt.Println()
}
