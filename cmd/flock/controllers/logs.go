package controllers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/GabeCordo/Flock/internal/core/component/message"
	"github.com/GabeCordo/Flock/internal/core/component/message/log"
	"github.com/GabeCordo/commandline"
)

type LogController struct {
}

func (controller LogController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	fileName := cli.NextArg()

	if fileName == commandline.FinalArg {

		// there is no log file specified, just output all the files

		err := filepath.Walk(DefaultLogsFolder, func(path string, info fs.FileInfo, err error) error {

			if (path == DefaultLogsFolder) || info.IsDir() {
				return nil
			}

			f, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			logs := strings.Split(string(f), "\n")
			numOfLogs := len(logs)

			fmt.Printf("├─ %s\t(num: %d)\n", info.Name(), numOfLogs)
			return nil
		})

		if err != nil {
			fmt.Print(err)
		}
	} else {

		// the operator specified a log file, scope into it
		priority := cli.NextArg()

		path := fmt.Sprintf("%s/%s", DefaultLogsFolder, fileName)
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("the log file does not exist (%s)\n", fileName)
		}

		logFile := log.NewLogFile(b)
		logFile.Print(message.FromShortform(priority))
	}

	return commandline.Terminate
}
