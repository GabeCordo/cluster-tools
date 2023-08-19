package database

import (
	"encoding/json"
	"errors"
	"github.com/GabeCordo/mango/components/cluster"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ConfigDatabase struct {
	records map[string]map[string]cluster.Config

	mutex sync.RWMutex
}

func NewConfigDatabase() *ConfigDatabase {

	db := new(ConfigDatabase)
	db.records = make(map[string]map[string]cluster.Config)

	return db
}

func (db *ConfigDatabase) Save(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path doesn't exist or isn't a directory")
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == curPath {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		os.RemoveAll(curPath)

		return nil
	})

	for moduleId, configs := range db.records {
		modulePath := path + moduleId
		if _, err := os.Stat(modulePath); err == nil {
			os.RemoveAll(modulePath)
		}
		os.Mkdir(modulePath, 0700)

		for identifier, config := range configs {
			configBytes, _ := json.Marshal(config)
			configPath := modulePath + "/" + identifier + ".json"
			f, _ := os.Create(configPath)
			f.Write(configBytes)
			f.Close()
		}
	}

	return nil
}

func (db *ConfigDatabase) Load(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path doesn't exist or isn't a directory")
	}

	filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if curPath == path {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		tmp := strings.Split(curPath, "/")
		moduleIdentifier := tmp[len(tmp)-2]

		fBytes, err := ioutil.ReadFile(curPath)
		if err != nil {
			return err
		}

		cfg := &cluster.Config{}
		if err = json.Unmarshal(fBytes, cfg); err != nil {
			return err
		}

		db.Create(moduleIdentifier, cfg.Identifier, *cfg)

		return nil
	})

	return nil
}

type ConfigFilter struct {
	Module     string
	Identifier string
}

func (db *ConfigDatabase) Get(filter ConfigFilter) (records []cluster.Config, err error) {

	records = nil
	err = nil

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if filter.Module == "" {
		err = errors.New("a module needs to at least be specified")
		return records, err
	}

	module, found := db.records[filter.Module]

	if !found {
		err = errors.New("module does not exist")
		return records, err
	}

	if filter.Identifier != "" {
		cnf, found := module[filter.Identifier]
		if !found {
			err = errors.New("no config with that identifier in this module")
			return records, err
		}
		records = make([]cluster.Config, 1)
		records[0] = cnf
	} else {
		records = make([]cluster.Config, len(module))

		idx := 0
		for _, cfg := range module {
			records[idx] = cfg
			idx++
		}
	}

	return records, err
}

func (db *ConfigDatabase) Create(moduleIdentifier, configIdentifier string, cfg cluster.Config) (err error) {

	err = nil

	db.mutex.Lock()
	defer db.mutex.Unlock()

	module, found := db.records[moduleIdentifier]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]cluster.Config)
		db.records[moduleIdentifier] = idToCfgMap
		module = idToCfgMap
	}

	_, found = module[configIdentifier]

	// if the config identifier already exists, we shouldn't be overwriting it
	// otherwise that can create unintended data side effects
	if found {
		err = errors.New("config with this identifier already exists in this module")
		return err
	}

	db.records[moduleIdentifier][configIdentifier] = cfg
	return err
}

func (db *ConfigDatabase) Replace(moduleIdentifier, configIdentifier string, cfg cluster.Config) (err error) {

	err = nil

	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, found := db.records[moduleIdentifier]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]cluster.Config)
		db.records[moduleIdentifier] = idToCfgMap
	}

	db.records[moduleIdentifier][configIdentifier] = cfg
	return err
}

func (db *ConfigDatabase) Delete(moduleIdentifier, configIdentifier string) (err error) {

	err = nil

	db.mutex.Lock()
	defer db.mutex.Unlock()

	configMap, found := db.records[moduleIdentifier]
	if !found {
		err = errors.New("module does not exist")
		return err
	}

	_, found = configMap[configIdentifier]
	if !found {
		err = errors.New("config does not exist")
		return err
	}

	delete(configMap, configIdentifier)

	return err
}
