package core

import (
	"fmt"
	"github.com/GabeCordo/fack"
	"gopkg.in/yaml.v3"
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
	config.Net.Port = 8000           // default
	config.Net.Host = fack.Localhost // default

	return config
}

func (config Config) ToYAML(path string) {

	// if a config already exists, delete it
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	file, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
	}
	_ = ioutil.WriteFile(path, file, DefaultFilePermissions)
}

func (config Config) Print() {

	bytes, _ := yaml.Marshal(config)
	fmt.Println(string(bytes))
}

func YAMLToETLConfig(config *Config, path string) error {
	if _, err := os.Stat(path); err != nil {
		// file does not exist
		log.Println(err)
		return err
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		// error reading the file
		log.Println(err)
		return err
	}

	err = yaml.Unmarshal([]byte(file), config)
	if err != nil {
		// the file is not a JSON or is a malformed (fields missing) config
		log.Println(err)
		return err
	}

	return nil
}
