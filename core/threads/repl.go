package threads

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"net/http"
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
		} else if text == "ping" {
			url := fmt.Sprintf("\"http://\"%s:%d/debug",
				common.GetConfigInstance().Net.Host, common.GetConfigInstance().Net.Port)
			request := &struct {
				Action string `json:"action"`
			}{
				Action: "ping",
			}
			body, _ := json.Marshal(request)
			http.Post(url, "application/json", bytes.NewReader(body))
		} else if text == "stop" {
			core.interrupt <- threads.Shutdown
			break
		}
	}
}
