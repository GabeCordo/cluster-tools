package controllers

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"

	"github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/fs/local"
)

type RunController struct {
}

func (controller RunController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.show(cli)
	} else if cli.Flag(commandline.Stop) {
		controller.stop(cli)
	} else {
		controller.create(cli)
	}

	return commandline.Terminate
}

func (controller RunController) create(cli *commandline.CommandLine) {

	pipeline := cli.NextArg()
	if pipeline == commandline.FinalArg {
		fmt.Println("you need to specify a pipeline to run")
		return
	}

	watch := cli.NextArg() == "--watch"

	core := local.GetCore()
	namespace := local.GetNamespace()

	fmt.Printf("trying to run %s.%s", namespace, pipeline)
	runId, err := api.RunPipelineOnCore(core, namespace, pipeline)
	if err != nil {
		fmt.Println("something went wrong :( is the core running?")
		return
	}

	if !watch {
		fmt.Printf(" -> track with id %d\n", runId)
	}

	sigs := make(chan os.Signal, 1)

	go func() {
		idx := 0
		for {
			time.Sleep(10 * time.Millisecond)

			run, err := api.GetRunStatus(core, namespace, runId)
			if err != nil {
				sigs <- syscall.SIGINT
			}

			if run.IsRunning() {
				if idx == 300 {
					idx = 0
					fmt.Print(".")
				}
				idx++
			} else {
				sigs <- syscall.SIGINT
			}
		}
	}()

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	<-sigs
}

func (controller RunController) show(cli *commandline.CommandLine) {

	idStr := cli.NextArg()
	if idStr == commandline.FinalArg {
		fmt.Println("you need to specify a run id")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("the run id must be a uint64")
		return
	}

	core := local.GetCore()
	namespace := local.GetNamespace()

	run, err := api.GetRunStatus(core, namespace, id)
	if err != nil {
		fmt.Println("something went wrong :(")
		return
	}
	fmt.Println(string(run.Status))
}

func (controller RunController) stop(cli *commandline.CommandLine) {

	idStr := cli.NextArg()
	if idStr == commandline.FinalArg {
		fmt.Println("you need to specify a run id")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("the run id must be a uint64")
		return
	}

	core := local.GetCore()

	err = api.StopRun(core, id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("triggered run stop")
	}
}
