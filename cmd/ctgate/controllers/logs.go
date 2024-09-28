package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/message"
	"github.com/GabeCordo/cluster-tools/internal/core/message/log"
	"github.com/GabeCordo/commandline"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type LogController struct {
}

func (controller LogController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	fileName := cli.NextArg()

	if fileName == commandline.FinalArg {

		// there is no log file specified, just output all the files

		filepath.Walk(DefaultLogsFolder, func(path string, info fs.FileInfo, err error) error {

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
