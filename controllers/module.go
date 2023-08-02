package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl-light/core/config"
	"github.com/GabeCordo/etl/framework/core"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
)

type ModuleCommand struct {
}

func (mc ModuleCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	workingDirectory, err := os.Getwd()
	if err != nil {
		panic("could not get working directory")
	}

	if cli.Flag(commandline.Show) {

	} else if cli.Flag(commandline.Install) {

		moduleYamlFile := workingDirectory + "/module.etl.yaml"
		if _, err := os.Stat(moduleYamlFile); err != nil {
			panic(err)
		}

		f, err := os.Open(moduleYamlFile)
		if err != nil {
			panic(err)
		}

		bytes, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}

		moduleConfig := config.Config{}
		if err := yaml.Unmarshal(bytes, &moduleConfig); err != nil {
			panic(err)
		}

		buildFolder := core.DefaultModulesFolder + moduleConfig.Name + "/"
		if _, err := os.Stat(buildFolder); err == nil {
			fmt.Printf("cannot create module %s as it already exists\n", moduleConfig.Name)
			return commandline.Terminate
		}

		os.Mkdir(buildFolder, 0700)

		buildCommand := exec.Command("go", "build", "-o", buildFolder)

		if err := buildCommand.Run(); err != nil {
			panic(err)
		}

		src, _ := os.Open(moduleYamlFile)
		defer src.Close()

		targetModuleYamlFile := buildFolder + "module.etl.yaml"
		if _, err := os.Stat(targetModuleYamlFile); err != nil {
			os.Create(targetModuleYamlFile)
		}

		dst, err := os.OpenFile(targetModuleYamlFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			panic(err)
		}

		dstModuleFolder := core.DefaultModulesFolder + "/" + moduleConfig.Name
		if _, err := os.Stat(dstModuleFolder); err == nil {
			os.Remove(dstModuleFolder)
		}

	} else if cli.Flag(commandline.Delete) {

		moduleName := cli.NextArg()
		if moduleName == commandline.FinalArg {
			fmt.Println("missing module name argument")
			return commandline.Terminate
		}

		buildFolder := core.DefaultModulesFolder + moduleName + "/"
		if _, err := os.Stat(buildFolder); os.IsNotExist(err) {
			fmt.Printf("%s module does not exist\n", moduleName)
			return commandline.Terminate
		}

		fmt.Printf("delete %s module (y/N)? ", moduleName)

		var option string
		fmt.Scanln(&option)

		if (option == "Y") || (option == "y") {
			os.RemoveAll(buildFolder)
			fmt.Printf("deleted module %s\n", moduleName)
		}
	}

	return commandline.Terminate
}
