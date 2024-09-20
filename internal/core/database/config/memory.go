package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type LocalConfigDatabase struct {
	records map[string]map[string]Config

	mutex sync.RWMutex
}

func NewLocalConfigDatabase() *LocalConfigDatabase {

	db := new(LocalConfigDatabase)
	db.records = make(map[string]map[string]Config)

	return db
}

func (db *LocalConfigDatabase) Save(path string) error {

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

func (db *LocalConfigDatabase) Load(path string) error {

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

		// HOTFIX: 2
		// this hotfix was put in place on 2023-11-03 by Gabriel Cordovado
		//
		// bug: org. had the separator as '/' which broke causing an index out of
		// 		bounds bug on windows devices.
		// fix: add the option for the \ separators on windows runtimes, also add check
		//		for index out of bounds in case future bugs arise.
		separator := "/"
		if runtime.GOOS == "windows" {
			separator = "\\"
		}
		tmp := strings.Split(curPath, separator)

		// if len is less than 2, then next operation would cause an index out of bounds
		if len(tmp) < 2 {
			log.Println("HOTFIX BUG! notify developers that there is another edge case for database separator")
			return nil
		}
		moduleIdentifier := tmp[len(tmp)-2]

		fBytes, err := ioutil.ReadFile(curPath)
		if err != nil {
			return err
		}

		cfg := &Config{}
		if err = json.Unmarshal(fBytes, cfg); err != nil {
			return err
		}

		db.Create(database.Filter{Module: moduleIdentifier, Cluster: cfg.Identifier}, cfg)

		return nil
	})

	return nil
}

type ConfigFilter struct {
	Module     string
	Identifier string
}

func (db *LocalConfigDatabase) Get(filter database.Filter) []any {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	results := make([]any, 0)

	if filter.Module == "" {
		return results
	}

	module, found := db.records[filter.Module]
	if !found {
		return results
	}

	if filter.Identifier != "" {
		cnf, found := module[filter.Identifier]
		if !found {
			return results
		}
		results = append(results, cnf)
	} else {
		for _, cfg := range module {
			results = append(results, cfg)
		}
	}

	return results
}

func (db *LocalConfigDatabase) Create(filter database.Filter, record any) (any, error) {

	cfg, ok := record.(*Config)
	if !ok {
		return nil, errors.New("LocalConfigDatabase expected *Config type")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	module, found := db.records[filter.Module]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]Config)
		db.records[filter.Module] = idToCfgMap
		module = idToCfgMap
	}

	_, found = module[cfg.Identifier]

	// if the config identifier already exists, we shouldn't be overwriting it
	// otherwise that can create unintended data side effects
	if found {
		return nil, errors.New("config with this identifier already exists in this module")
	}

	db.records[filter.Module][cfg.Identifier] = *cfg // copy
	return cfg.Identifier, nil
}

func (db *LocalConfigDatabase) Replace(filter database.Filter, record any) error {

	cfg, ok := record.(*Config)
	if !ok {
		return errors.New("LocalConfigDatabase expected *Config type")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, found := db.records[filter.Module]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]Config)
		db.records[filter.Module] = idToCfgMap
	}

	db.records[filter.Module][cfg.Identifier] = *cfg
	return nil
}

func (db *LocalConfigDatabase) Delete(filter database.Filter) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	configMap, found := db.records[filter.Module]
	if !found {
		return errors.New("module does not exist")
	}

	_, found = configMap[filter.Identifier]
	if !found {
		return errors.New("config does not exist")
	}

	delete(configMap, filter.Identifier)
	return nil
}

func (db *LocalConfigDatabase) Print() {

	for moduleName, module := range db.records {

		fmt.Printf("├─ %s\n", moduleName)

		for clusterName, _ := range module {
			fmt.Printf("|   ├─ %s\n", clusterName)
		}
	}
}
