package core

import (
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/fack"
	"io/ioutil"
	"log"
	"os"
)

const (
	DefaultFilePermissions os.FileMode = 0755
)

func NewConfig(name string) *Config {
	config := new(Config)

	config.Name = name
	config.Version = 1.0
	config.Net.Port(8080)
	config.Net.Host(fack.Localhost)

	return config
}

func (config Config) ToJson(path string) {

	// if a config already exists, delete it
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	_ = ioutil.WriteFile(path, file, DefaultFilePermissions)
}

func JSONToETLConfig(config *Config, path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		// file does not exist
		log.Println(err)
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
