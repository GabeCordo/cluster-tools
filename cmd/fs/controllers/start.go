package controllers

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/FortifiedCode/commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core"
)

type StartCommand struct {
}

func (sc StartCommand) banner() {
	fmt.Println("   __ _            _    \n  / _| | ___   ___| | __\n | |_| |/ _ \\ / __| |/ /\n |  _| | (_) | (__|   < \n |_| |_|\\___/ \\___|_|\\_\\")
	fmt.Println("[+] " + terminal.Purple + "The Distributed Service Framework " + terminal.Reset)
	fmt.Println("[+]" + terminal.Purple + " by Gabriel Cordovado 2022-25" + terminal.Reset)
	fmt.Println()
}

const MongoDatabaseUriEnv = "MONGO_DATABASE_URI"

type EnvironmentVariables struct {
	MongoDbUri string
}

func (sc StartCommand) readEnvironmentVariables() (env EnvironmentVariables) {

	env.MongoDbUri = os.Getenv(MongoDatabaseUriEnv)
	return env
}

func (sc StartCommand) verifyMandatoryEnvironmentVariables(env EnvironmentVariables) (err error) {

	if env.MongoDbUri == "" {
		output := fmt.Sprintf("the environment variable %s needs to be set", MongoDatabaseUriEnv)
		err = errors.New(output)
	}

	return err
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// check to see that the etl thread has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("missing configurations folder at %s\nmake sure you run 'DistributedFunctions init'\n", DefaultConfigsFolder)
		return commandline.Terminate
	}

	sc.banner()

	env := sc.readEnvironmentVariables()
	err := sc.verifyMandatoryEnvironmentVariables(env)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	cfg, err := core.GetConfigInstance(DefaultCoreConfigFile)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}
	cfg.Database.Url = env.MongoDbUri

	c, err := core.New(cfg)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	c.Run()

	return commandline.Terminate
}
