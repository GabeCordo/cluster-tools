package in_memory

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

type LocalDatabase struct {
	records map[string]map[string]*plover.PipelineIR

	mutex sync.RWMutex
}

func NewLocalPipelineDatabase() *LocalDatabase {

	localDatabase := new(LocalDatabase)
	localDatabase.records = make(map[string]map[string]*plover.PipelineIR)

	return localDatabase
}

// Save writes records from the pipeline.LocalDatabase to disk.
func (localDatabase *LocalDatabase) Save(path string) error {

	path = filepath.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path doesn't exist or isn't a directory")
	}

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

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

	for moduleId, configs := range localDatabase.records {
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

// Load reads records from the local disk to pipeline.LocalDatabase.
func (localDatabase *LocalDatabase) Load(path string) error {

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

		_, err = localDatabase.Create(database.Filter{Namespace: moduleIdentifier, Pipeline: cfg.Identifier}, cfg)
		return err
	})

	return err
}

// Get retrieves a *plover.PipelineIR record from the pipeline.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) (results []*plover.PipelineIR) {

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	if filter.Namespace == "" {
		return results
	}

	module, found := localDatabase.records[filter.Namespace]
	if !found {
		return results
	}

	var pipeline *plover.PipelineIR
	if filter.Identifier != "" {
		pipeline, found = module[filter.Identifier]
		if !found {
			return results
		}
		results = append(results, pipeline)
	} else {
		for _, cfg := range module {
			results = append(results, cfg)
		}
	}

	return results
}

// Create adds a new *plover.PipelineIR record to the pipeline.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, record *plover.PipelineIR) (string, error) {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	module, found := localDatabase.records[filter.Namespace]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]*plover.PipelineIR)
		localDatabase.records[filter.Namespace] = idToCfgMap
		module = idToCfgMap
	}

	_, found = module[record.Identifier]

	// if the pipeline identifier already exists, we shouldn't be overwriting it
	// otherwise that can create unintended data side effects
	if found {
		return "", errors.New("pipeline with this identifier already exists in this module")
	}

	localDatabase.records[filter.Namespace][record.Identifier] = record
	return record.Identifier, nil
}

// Replace swaps a *plover.PipelineIR with an existing record in the pipeline.LocalDatabase.
func (localDatabase *LocalDatabase) Replace(filter database.Filter, record *plover.PipelineIR) error {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	_, found := localDatabase.records[filter.Namespace]

	// the module needs to exist for us to add new configs to it
	// if it doesn't exist, lazily create it in the database
	if !found {
		idToCfgMap := make(map[string]*plover.PipelineIR)
		localDatabase.records[filter.Namespace] = idToCfgMap
	}

	localDatabase.records[filter.Namespace][record.Identifier] = record
	return nil
}

// Delete removes a *plover.PipelineIR from the pipeline.LocalDatabase.
func (localDatabase *LocalDatabase) Delete(filter database.Filter) error {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	configMap, found := localDatabase.records[filter.Namespace]
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

// Print outputs the *plover.PipelineIR records in the pipeline.LocalDatabase to the console.
func (localDatabase *LocalDatabase) Print() {

	for moduleName, module := range localDatabase.records {

		fmt.Printf("├─ %s\n", moduleName)

		for clusterName := range module {
			fmt.Printf("|   ├─ %s\n", clusterName)
		}
	}
}
