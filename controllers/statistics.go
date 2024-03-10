package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/commandline"
	"io/fs"
	"path/filepath"
)

type StatisticsController struct {
}

func (controller StatisticsController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	filepath.Walk(common.DefaultStatisticsFolder, func(path string, info fs.FileInfo, err error) error {

		if (path == common.DefaultLogsFolder) || info.IsDir() {
			return nil
		}

		fmt.Printf("├─ %s (bytes: %d)\n", info.Name(), info.Size())
		return nil
	})

	return commandline.Terminate
}
