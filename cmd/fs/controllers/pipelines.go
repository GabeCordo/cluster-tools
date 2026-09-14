package controllers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"

	"github.com/FortifiedCode/commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/DistributedFunctions/local"
	"gopkg.in/yaml.v3"
)

type PipelineController struct {
}

func (controller PipelineController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	rootPath := cli.NextArg()
	if rootPath == commandline.FinalArg {
		log.Println("missing pipeline path file")
		return commandline.Terminate
	}

	// is the provided path an absolute path?
	// todo: maybe use a os.stat check instead of this dot trick
	var providedPath string
	if rootPath[0:1] == "." {
		workingDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}

		providedPath = filepath.Join(workingDirectory, rootPath)
	} else {
		providedPath = rootPath
	}

	fInfo, err := os.Stat(providedPath)
	if err != nil {
		log.Printf("the path you provided doesn't seem to exist? %s\n", providedPath)
		return commandline.Terminate
	}

	pipelinePaths := make([]string, 0)
	pipelines := make([]*ScalingFunctions.PipelineIR, 0)

	// if the path provided is a folder, try to grab all the deployment (pipeline) files
	// inside the folder and register them on the core
	if fInfo.IsDir() {
		err := filepath.Walk(providedPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || err != nil {
				return nil
			}
			pipelinePaths = append(pipelinePaths, path)
			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
	} else {
		pipelinePaths = append(pipelinePaths, rootPath)
	}

	// pull all the pipline files from the paths we have collected
	for _, pipelinePath := range pipelinePaths {

		pipelinePath = filepath.Clean(pipelinePath)
		f, err := os.Open(pipelinePath)
		if err != nil {
			fmt.Println(err)
			// TODO : add better error outputs
			continue
		}

		p := &struct {
			Pipeline *ScalingFunctions.PipelineIR `yaml:"pipeline"`
		}{}
		if err = yaml.NewDecoder(f).Decode(p); err == nil {
			pipelines = append(pipelines, p.Pipeline)
		} else {
			fmt.Println(err)
		}

		err = f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}

	if cli.Flag(commandline.Add) {
		controller.addPipelines(pipelines)
	} else if cli.Flag(commandline.Delete) {
		controller.deletePipelines(pipelines)
	} else {
		fmt.Println("no operation specified.")
	}

	return commandline.Terminate
}

func (controller PipelineController) addPipelines(pipelines []*ScalingFunctions.PipelineIR) {

	core := local.GetCore()
	namespace := local.GetNamespace()

	// register each pipeline with the gateway
	for i, p := range pipelines {

		fmt.Printf("[%d] adding pipleine %s ", i, p.Identifier)
		onCore, err := api.IsPipelineOnCore(core, namespace, p.Identifier)
		if err != nil {
			fmt.Println("...skip")
			continue
		}

		if !onCore {
			if err = api.CreatePipelineOnCore(core, namespace, p); err != nil {
				fmt.Println("... failed")
			} else {
				fmt.Println("... done")
			}
		} else {
			fmt.Println("... already exists")
		}

	}
}

func (controller PipelineController) deletePipelines(pipelines []*ScalingFunctions.PipelineIR) {

	core := local.GetCore()
	namespace := local.GetNamespace()

	for i, p := range pipelines {
		fmt.Printf("[%d] removing pipeline %s...", i, p.Identifier)
		if err := api.RemovePipelineFromCore(core, namespace, p.Identifier); err == nil {
			fmt.Println("... done")
		} else {
			fmt.Println("... skip")
		}
	}
}
