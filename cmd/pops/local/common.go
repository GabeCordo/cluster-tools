package local

import (
	"encoding/json"
	"os"
)

type Config struct {
	Namespace string `json:"namespace"`
	Core      string `json:"core"`
}

var (
	userCacheDir, _        = os.UserCacheDir()
	DefaultFrameworkFolder = userCacheDir + "/PipelineOps/"
	CToolsFolder           = DefaultFrameworkFolder + "/ctools/"
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
	defer f.Close()

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
	defer f.Close()

	return json.NewEncoder(f).Encode(config)
}

func getConfig(config *Config) error {

	f, err := os.Open(CToolsConfig)
	if err != nil {
		panic(err)
	}
	defer f.Close()

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
