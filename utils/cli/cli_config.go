package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	DefaultJSONPrefix string = ""
	DefaultJSONIndent string = " "
)

func (config *Config) ToJson() {
	executablePath, _ := os.Executable()
	defaultConfigPath := executablePath[:len(executablePath)-9] + "config.cli.json"

	// if a config already exists, delete it
	if _, err := os.Stat(defaultConfigPath); err == nil {
		os.Remove(defaultConfigPath)
	}

	file, err := json.MarshalIndent(config, DefaultJSONPrefix, DefaultJSONIndent)
	if err != nil {
		fmt.Println(err)
	}
	_ = ioutil.WriteFile(defaultConfigPath, file, DefaultFilePermissions)
}

func JSONToCLIConfig(config *Config) error {
	executablePath, _ := os.Executable()

	// check to see if the executable is being run in a goland debugger
	if executablePath[1:8] == "private" {
		executablePath = "/Users/gabecordovado/go/src/etl/"
	} else {
		executablePath = executablePath[:len(executablePath)-9]
	}
	defaultConfigPath := executablePath + "config.cli.json"

	file, err := ioutil.ReadFile(defaultConfigPath)
	if err != nil {
		// cannot proceed without the cli config
		panic("missing etl cli config")
		return err
	}

	err = json.Unmarshal([]byte(file), config)
	if err != nil {
		// the file is not a JSON or is a malformed (fields missing) config
		log.Println(err)
		return err
	}

	return nil
}
