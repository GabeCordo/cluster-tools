package core

import (
	"fmt"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
)

func (core *Core) banner() {
	fmt.Println("   __ _            _    \n  / _| | ___   ___| | __\n | |_| |/ _ \\ / __| |/ /\n |  _| | (_) | (__|   < \n |_| |_|\\___/ \\___|_|\\_\\")
	fmt.Println("[+] " + terminal.Purple + "The Distributed Service Framework " + terminal.Reset + Version)
	fmt.Println("[+]" + terminal.Purple + " by Gabriel Cordovado 2022-25" + terminal.Reset)
	fmt.Println()
}
