package client

import (
	"encoding/json"
	"github.com/GabeCordo/commandline"
	"time"
)

type Project struct {
	Name      string    `json:"PublicName"`
	Directory string    `json:"directory"`
	CreatedOn time.Time `json:"created-on"`
}

type Config struct {
	Projects []Project `json:"projects"`
}

func NewConfig() *Config {
	config := new(Config)
	config.Projects = make([]Project, 0)

	return config
}

func (config *Config) AddProject(project Project) {
	config.Projects = append(config.Projects, project)
}

func (config *Config) ToJson(path commandline.Path) error {
	if path.DoesNotExist() {
		panic("the path is not valid, it cannot be converted to JSON")
	}

	bytes, err := json.MarshalIndent(config, commandline.DefaultJSONPrefix, commandline.DefaultJSONIndent)
	if err != nil {
		panic("there was an issue marshalling the Config to JSON")
	}

	return path.Write(bytes)
}

func JSONToConfig(path commandline.Path) *Config {
	if path.DoesNotExist() {
		panic("the path provided for the JSON file " + path.ToString() + " does not exist")
	}

	bytes, err := path.Read()
	if err != nil {
		panic("there was an error while reading the JSON file " + path.ToString())
	}

	config := NewConfig()

	err = json.Unmarshal(bytes, config)
	if err != nil {
		panic("there was an issue unmarshalling JSON into client.Config")
	}

	return config
}
