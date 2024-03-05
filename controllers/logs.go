package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
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

		filepath.Walk(common.DefaultLogsFolder, func(path string, info fs.FileInfo, err error) error {

			if (path == common.DefaultLogsFolder) || info.IsDir() {
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

		path := fmt.Sprintf("%s/%s", common.DefaultLogsFolder, fileName)
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("the log file does not exist (%s)\n", fileName)
		}

		logFile := messenger.NewLogFile(b)
		logFile.Print(messenger.MessagePriority(priority))
	}

	return commandline.Terminate
}
