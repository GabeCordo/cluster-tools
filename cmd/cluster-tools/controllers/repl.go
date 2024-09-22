package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
	"os"
)

type ReplController struct {
}

func (controller ReplController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	option := cli.NextArg()
	if option == commandline.FinalArg {
		fmt.Println("when using the 'repl' command expected [enable|disable] to modify what happens on 'mango start'")
	}

	//// excerpt from : https://stackoverflow.com/questions/62000607/how-to-overwrite-file-content-in-golang
	//configFile, err := os.OpenFile(common.DefaultConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	//if err != nil {
	//	fmt.Printf("[x] cannot find mango pipeline(%s). Is it possible 'mango init' was never called?\n",
	//		common.DefaultConfigFile)
	//	return commandline.Terminate
	//}
	//defer configFile.Close()

	bytes, err := os.ReadFile(DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	}

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global common is corrupt (%s)\n", DefaultConfigFile)
		return commandline.Terminate
	}

	if option == "enable" {
		c.EnableRepl = true
	} else {
		c.EnableRepl = false
	}

	bytes, err = yaml.Marshal(c)
	if err != nil {
		fmt.Printf("[x] failed to modify the pipeline file at %s\n", DefaultConfigFile)
		return commandline.Terminate
	}

	err = os.WriteFile(DefaultConfigFile, bytes, 0755)
	if err != nil {
		fmt.Printf("[x] failed to modify the pipeline file at %s\n", DefaultConfigFile)
		return commandline.Terminate
	}

	return commandline.Terminate
}
