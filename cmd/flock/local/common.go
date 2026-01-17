package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const defaultFilePerm = 0777

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

	f, err := os.OpenFile(ToolsConfig, os.O_RDWR|os.O_CREATE, defaultFilePerm) // #nosec G304 -- Constant is not user controlled
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

	f, err := os.OpenFile(ToolsConfig, os.O_RDWR, defaultFilePerm) // #nosec G304 -- Constant is not user controlled
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

func getConfig(config *Config) (err error) {

	if config == nil {
		return errors.New("config is nil")
	}

	var f *os.File
	f, err = os.Open(ToolsConfig) // #nosec G304 -- Constant is not user controlled
	if err != nil {
		return err
	}

	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return err
	}

	return f.Close()
}

func getOrCreateConfig(config *Config) (err error) {

	if config == nil {
		return errors.New("config is nil")
	}

	err = getConfig(config)
	if err != nil {
		config.Namespace = "common"
		config.Core = "http://localhost:8136"

		err = createConfig(config)
	}

	return err
}
