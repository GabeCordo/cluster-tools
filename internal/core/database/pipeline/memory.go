package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FortifiedCode/plover"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/FortifiedCode/flock/internal/core/database"
)

type LocalPipelineDatabase struct {
	records map[string]map[string]plover.PipelineIR

	mutex sync.RWMutex
}

func NewLocalPipelineDatabase() *LocalPipelineDatabase {

	db := new(LocalPipelineDatabase)
	db.records = make(map[string]map[string]plover.PipelineIR)

	return db
}

func (db *LocalPipelineDatabase) Save(path string) error {

	path = filepath.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path doesn't exist or isn't a directory")
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	err := filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == curPath {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		err = os.RemoveAll(curPath)
		return err
	})

	if err != nil {
		return err
	}

	for moduleId, configs := range db.records {
		modulePath := filepath.Join(path, moduleId)

		if _, err := os.Stat(modulePath); err == nil {
			err = os.RemoveAll(modulePath)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		err := os.Mkdir(modulePath, 0700)
		if err != nil {
			log.Println(err)
			continue
		}

		for identifier, config := range configs {
			configBytes, _ := json.Marshal(config)
			fileName := fmt.Sprintf("%s.json", identifier)
			configPath := filepath.Join(modulePath, fileName)
			configPath = filepath.Clean(configPath)
			f, err := os.Create(configPath)
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = f.Write(configBytes)
			if err != nil {
				log.Println(err)
				continue
			}
			err = f.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}

func (db *LocalPipelineDatabase) Load(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path doesn't exist or isn't a directory")
	}

	err := filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
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

		curPath = filepath.Clean(curPath)
		f, err := os.Open(curPath)
		if err != nil {
			return err
		}

		cfg := &plover.PipelineIR{}
		if err = json.NewDecoder(f).Decode(cfg); err != nil {
			return err
		}

		_, err = db.Create(database.Filter{Namespace: moduleIdentifier, Pipeline: cfg.Identifier}, cfg)
		return err
	})

	return err
}

type ConfigFilter struct {
	Module     string
	Identifier string
}

func (db *LocalPipelineDatabase) Get(filter database.Filter) []any {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	results := make([]any, 0)

	if filter.Namespace == "" {
		return results
	}

	module, found := db.records[filter.Namespace]
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

func (db *LocalPipelineDatabase) Create(filter database.Filter, record any) (any, error) {

	cfg, ok := record.(*plover.PipelineIR)
	if !ok {
		return nil, errors.New("LocalPipelineDatabase expected *pipeline type")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	module, found := db.records[filter.Namespace]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]plover.PipelineIR)
		db.records[filter.Namespace] = idToCfgMap
		module = idToCfgMap
	}

	_, found = module[cfg.Identifier]

	// if the pipeline identifier already exists, we shouldn't be overwriting it
	// otherwise that can create unintended data side effects
	if found {
		return nil, errors.New("pipeline with this identifier already exists in this module")
	}

	db.records[filter.Namespace][cfg.Identifier] = *cfg // copy
	return cfg.Identifier, nil
}

func (db *LocalPipelineDatabase) Replace(filter database.Filter, record any) error {

	cfg, ok := record.(*plover.PipelineIR)
	if !ok {
		return errors.New("LocalPipelineDatabase expected *pipeline type")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, found := db.records[filter.Namespace]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]plover.PipelineIR)
		db.records[filter.Namespace] = idToCfgMap
	}

	db.records[filter.Namespace][cfg.Identifier] = *cfg
	return nil
}

func (db *LocalPipelineDatabase) Delete(filter database.Filter) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	configMap, found := db.records[filter.Namespace]
	if !found {
		return errors.New("module does not exist")
	}

	_, found = configMap[filter.Identifier]
	if !found {
		return errors.New("pipeline does not exist")
	}

	delete(configMap, filter.Identifier)
	return nil
}

func (db *LocalPipelineDatabase) Print() {

	for moduleName, module := range db.records {

		fmt.Printf("├─ %s\n", moduleName)

		for clusterName := range module {
			fmt.Printf("|   ├─ %s\n", clusterName)
		}
	}
}
