package local

import (
	"encoding/json"
	"fmt"
	"os"
)

const defaultFilePerm = 0600

type Config struct {
	Namespace string `json:"namespace"`
	Core      string `json:"core"`
}

var (
	userCacheDir, _        = os.UserCacheDir()
	DefaultFrameworkFolder = userCacheDir + "/flock/"
	ToolsFolder            = DefaultFrameworkFolder + "/tools/"
	ToolsConfig            = ToolsFolder + "config.json"
)

func createConfig(config *Config) error {

	if _, err := os.Stat(userCacheDir); os.IsNotExist(err) {
		if err = os.MkdirAll(DefaultFrameworkFolder, defaultFilePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(ToolsFolder); os.IsNotExist(err) {
		if err = os.MkdirAll(ToolsFolder, defaultFilePerm); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(ToolsConfig, os.O_RDWR|os.O_CREATE, defaultFilePerm)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(f)

	if err = json.NewEncoder(f).Encode(config); err != nil {
		return err
	}
	return nil
}

func updateConfig(config *Config) error {

	err := os.Truncate(ToolsConfig, 0)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(ToolsConfig, os.O_RDWR, defaultFilePerm)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(f)

	return json.NewEncoder(f).Encode(config)
}

func getConfig(config *Config) error {

	f, err := os.Open(ToolsConfig)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(f)

	return json.NewDecoder(f).Decode(config)
}

func getOrCreateConfig(config *Config) error {

	if err := getConfig(config); err != nil {

		config.Namespace = "common"
		config.Core = "http://localhost:8136"

		if err = createConfig(config); err != nil {
			return err
		}
	}

	return nil
}
