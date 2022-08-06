package core

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

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
