package local

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Namespace string `json:"namespace"`
	Core      string `json:"core"`
}

var (
	userCacheDir, _        = os.UserCacheDir()
	DefaultFrameworkFolder = userCacheDir + "/flock/"
	CToolsFolder           = DefaultFrameworkFolder + "/tools/"
	CToolsConfig           = CToolsFolder + "config.json"
)

func createConfig(config *Config) error {

	if _, err := os.Stat(userCacheDir); os.IsNotExist(err) {
		if err = os.MkdirAll(DefaultFrameworkFolder, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(CToolsFolder); os.IsNotExist(err) {
		if err = os.MkdirAll(CToolsFolder, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(CToolsConfig, os.O_RDWR|os.O_CREATE, 0755)
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

	err := os.Truncate(CToolsConfig, 0)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(CToolsConfig, os.O_RDWR, 0755)
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

	f, err := os.Open(CToolsConfig)
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
