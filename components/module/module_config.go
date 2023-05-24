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
	return true
}
