package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/cmd/ctools/api"
	"github.com/GabeCordo/cluster-tools/cmd/ctools/local"
	"github.com/GabeCordo/commandline"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type StatsController struct {
}

func (controller StatsController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	}

	return commandline.Terminate
}

func (controller StatsController) showFlag(cli *commandline.CommandLine) {

	idStr := cli.NextArg()
	if idStr == commandline.FinalArg {
		fmt.Println("missing identifier of run")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("the identifier of a run must be uint64")
		return
	}

	core := local.GetCore()
	namespace := local.GetNamespace()

	sigs := make(chan os.Signal, 1)

	go func() {

		for {

			run, err := api.GetRunStatus(core, namespace, id)
			if err != nil {
				fmt.Println("there was an issue fetching the statistics of a run")
				return
			}

			if !run.IsRunning() {
				fmt.Println("=== done ===")
				sigs <- syscall.SIGTERM
			}

			numOfGoroutines := 0
			for _, f := range run.Statistics.Functions {
				numOfGoroutines += f.Active
			}
			fmt.Printf("running functions: %d\n", numOfGoroutines)

			time.Sleep(100 * time.Millisecond)
		}
	}()

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	<-sigs
}
