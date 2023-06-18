package module

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
)

func NewRemoteModule(path string) (*RemoteModule, error) {

	remoteModule := new(RemoteModule)
	remoteModule.Path = path

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	return remoteModule, nil
}

func (remoteModule RemoteModule) Get() (*Module, error) {

	module := new(Module)

	filepath.Walk(remoteModule.Path, func(path string, info os.FileInfo, err error) error {

		if info.Name() == "module.etl.yaml" {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				log.Println(err)
				return err
			}

			module.Config = &Config{}
			if err = yaml.Unmarshal(bytes, module.Config); err != nil {
				log.Println(err)
				return err
			}
		} else if filepath.Ext(info.Name()) == ".so" {
			if module.Plugin, err = plugin.Open(path); err != nil {
				log.Println(err)
				return err
			}
		}

		return nil
	})

	if (module.Plugin == nil) || (module.Config == nil) {
		return nil, os.ErrExist
	}

	return module, nil
}
