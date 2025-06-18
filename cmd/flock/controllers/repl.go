package controllers

import (
	"fmt"
	"os"

	"github.com/GabeCordo/Flock/internal/core"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
)

type ReplController struct {
}

func (controller ReplController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	option := cli.NextArg()
	if option == commandline.FinalArg {
		fmt.Println("when using the 'repl' command expected [enable|disable] to modify what happens on 'mango start'")
	}

	//// excerpt from : https://stackoverflow.com/questions/62000607/how-to-overwrite-file-content-in-golang
	//configFile, err := os.OpenFile(common.DefaultCoreConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	//if err != nil {
	//	fmt.Printf("[x] cannot find mango pipeline(%s). Is it possible 'mango init' was never called?\n",
	//		common.DefaultCoreConfigFile)
	//	return commandline.Terminate
	//}
	//defer configFile.Close()

	bytes, err := os.ReadFile(DefaultCoreConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultCoreConfigFile)
		return commandline.Terminate
	}

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultCoreConfigFile)
		return commandline.Terminate
	}

	if option == "enable" {
		c.EnableRepl = true
	} else {
		c.EnableRepl = false
	}

	bytes, err = yaml.Marshal(c)
	if err != nil {
		fmt.Printf("[x] failed to modify the pipeline file at %s\n", DefaultCoreConfigFile)
		return commandline.Terminate
	}

	err = os.WriteFile(DefaultCoreConfigFile, bytes, defaultFilePerm)
	if err != nil {
		fmt.Printf("[x] failed to modify the pipeline file at %s\n", DefaultCoreConfigFile)
		return commandline.Terminate
	}

	return commandline.Terminate
}
