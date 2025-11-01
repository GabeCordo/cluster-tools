package core

import (
	"bufio"
	"fmt"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	"os"
	"strings"
)

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
			core.interrupt <- thread.Shutdown
			break
		}
	}
}
