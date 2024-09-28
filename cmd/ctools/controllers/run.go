package controllers

import (
	"github.com/GabeCordo/cluster-tools/cmd/ctools/api"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type RunCommand struct {
}

func (controller RunCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	gatway, err := api.Gateway()
	if err != nil {
		log.Println("no processor is running? have you started it.")
		return commandline.Terminate
	}

	if len(gatway) == 0 {
		// the gateway is running in standalone mode

		pipelinePath := cli.NextArg()
		if pipelinePath == commandline.FinalArg {
			log.Println("missing pipeline path file")
			return commandline.Terminate
		}

		f, err := os.Open(pipelinePath)
		if err != nil {
			log.Println("something went wrong wile opening the pipeline file")
			return commandline.Terminate
		}

		wrapper := &struct {
			Pipeline *pipeline.Pipeline `yaml:"pipeline"`
		}{}

		if err = yaml.NewDecoder(f).Decode(wrapper); err != nil {
			log.Println("is this a pipeline file? something went wrong while opening the pipeline file")
			return commandline.Terminate
		}

		if err = api.RunPipelineOnProcessor(wrapper.Pipeline); err != nil {
			log.Println("could not run pipeline on processor")
		}

	} else {
		// the gateway is running in connected mode
	}

	return commandline.Terminate
}
