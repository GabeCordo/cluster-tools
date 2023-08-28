package module

import (
	"gopkg.in/yaml.v3"
	"os"
)

func ConfigFromYAML(path string) (*Config, error) {

	config := new(Config)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(bytes, config)

	return config, nil
}

func (config Config) Verify() bool {

	// ensure that every export identifier is unique
	exports := make(map[string]bool)
	for _, export := range config.Exports {
		if _, found := exports[export.Cluster]; found {
			return false
		} else {
			exports[export.Cluster] = true
		}
	}

	return true
}
