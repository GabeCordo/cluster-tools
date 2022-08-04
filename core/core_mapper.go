package core

import (
	"encoding/json"
	"io/ioutil"
)

func JSONToETLConfig(config *Config, path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		// file does not exist
		return err
	}

	err = json.Unmarshal([]byte(file), config)
	if err != nil {
		// the file is not a JSON or is a malformed (fields missing) config
		return err
	}

	return nil
}
