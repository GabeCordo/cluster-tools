package core

import (
	"bufio"
	"fmt"
	"github.com/GabeCordo/mango/core/threads/common"
	"os"
	"strings"
)

func banner() {
	fmt.Println()
	fmt.Println("the interactive shell is an experimental feature that is still being worked on. " +
		"there may be some issues or missing features that are under development.")
	fmt.Println()
}

func (core *Core) repl() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("@etl ")
		text, _ := reader.ReadString('\n')
		text = strings.ReplaceAll(text, "\n", "")

		if text == "modules" {
			//p := GetProvisionerInstance()
			//modules := p.GetModules()
			//
			//for _, module := range modules {
			//	module.Print()
			//}
			fmt.Println("not implemented")
		} else if text == "stop" {
			core.interrupt <- common.Shutdown
			break
		}
	}
}
