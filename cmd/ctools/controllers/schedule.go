package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/cmd/ctools/api"
	"github.com/GabeCordo/cluster-tools/cmd/ctools/local"
	"github.com/GabeCordo/cluster-tools/internal/core/database"
	"github.com/GabeCordo/cluster-tools/internal/core/database/job"
	"github.com/GabeCordo/commandline"
	"strconv"
)

type ScheduleController struct {
}

func (controller ScheduleController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	} else if cli.Flag(commandline.Add) {
		controller.addFlag(cli)
	} else if cli.Flag(commandline.Delete) {
		controller.deleteFlag(cli)
	}

	return commandline.Terminate
}

func (controller ScheduleController) showFlag(cli *commandline.CommandLine) {

	core := local.GetCore()
	namespace := local.GetNamespace()

	jj, err := api.GetJobs(core, namespace)
	if err != nil {
		fmt.Println("is the core running?")
		return
	}

	for _, j := range jj {
		output := j.ToString()
		fmt.Println(output)
	}
}

func (controller ScheduleController) addFlag(cli *commandline.CommandLine) {

	id := cli.NextArg()
	if id == commandline.FinalArg {
		fmt.Println("missing an identifier for the job")
		return
	}

	pipeline := cli.NextArg()
	if pipeline == commandline.FinalArg {
		fmt.Println("missing a pipeline for the job")
		return
	}

	intervalStr := cli.NextArg()
	if intervalStr == commandline.FinalArg {
		fmt.Println("missing an interval for the job to run on")
		return
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		fmt.Println("the interval must be a valid uint")
		return
	}

	if (interval > 60) || (interval < 1) {
		fmt.Println("the interval must be in the range 1-60 minutes")
		return
	}

	namespace := local.GetNamespace()
	core := local.GetCore()

	j := job.Job{
		Identifier: id,
		Namespace:  namespace,
		Pipeline:   pipeline,
		Interval:   database.Interval{Minute: interval},
		Metadata:   make(map[string]string),
	}

	if err = api.CreateJob(core, j); err != nil {
		fmt.Println("is the core running?")
	}
}

func (controller ScheduleController) deleteFlag(cli *commandline.CommandLine) {

	id := cli.NextArg()
	if id == commandline.FinalArg {
		fmt.Println("missing the id of the scheduled job")
		return
	}

	core := local.GetCore()

	if err := api.DeleteJob(core, id); err != nil {
		fmt.Println("is the core running?")
	}
}
