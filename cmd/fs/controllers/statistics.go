package controllers

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/GabeCordo/Commandline"
)

type StatisticsController struct {
}

func (controller StatisticsController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	err := filepath.Walk(DefaultStatisticsFolder, func(path string, info fs.FileInfo, err error) error {

		if (path == DefaultLogsFolder) || info.IsDir() {
			return nil
		}

		fmt.Printf("├─ %s (bytes: %d)\n", info.Name(), info.Size())
		return nil
	})

	if err != nil {
		fmt.Print(err)
	}

	return commandline.Terminate
}
