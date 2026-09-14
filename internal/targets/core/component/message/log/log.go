package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/message"
)

var logRegex = regexp.MustCompile(`\[(.+)]\[(.+)](.+)`)

type Log struct {
	Id        uint64
	Timestamp time.Time
	Priority  message.Priority
	Message   string
}

func NewLog() *Log {
	log := new(Log)
	return log
}

func (log *Log) ToString() string {
	return fmt.Sprintf("[%s][%s] %s", time.Now().Format(time.RFC3339Nano), log.Priority.Shortform(), log.Message)
}

func (log *Log) Parse(data string) error {

	matches := logRegex.FindStringSubmatch(data)
	if len(matches) != 4 {
		return errors.New("invalid log string")
	}

	t, err := time.Parse(time.RFC3339Nano, matches[1])
	if err == nil {
		return errors.New("log has invalid time format")
	}

	log.Timestamp = t
	log.Priority = message.FromShortform(matches[2])
	log.Message = matches[3]

	return nil
}

type Cluster struct {
	supervisors map[uint64][]string
	mutex       sync.RWMutex
}

func NewCluster() *Cluster {
	instance := new(Cluster)
	instance.supervisors = make(map[uint64][]string)
	return instance
}

type Module struct {
	clusters map[string]*Cluster
	mutex    sync.RWMutex
}

func NewModule() *Module {
	instance := new(Module)
	instance.clusters = make(map[string]*Cluster)
	return instance
}

type Logger struct {
	enabled struct {
		logging bool
		smtp    bool
	}
	directory string

	modules map[string]*Module
	mutex   sync.RWMutex
}

func New(enableLogging bool) *Logger {

	logger := new(Logger)
	logger.modules = make(map[string]*Module)
	logger.enabled.logging = enableLogging

	return logger
}

func (logger *Logger) LoggingDirectory(path string) *Logger {
	if logger.enabled.logging {
		logger.directory = path
	}

	return logger
}

func (logger *Logger) Message(source message.Source, record any) error {

	logger.mutex.Lock()

	moduleInstance, moduleFound := logger.modules[source.Module]

	if !moduleFound {
		moduleInstance = NewModule()
		logger.modules[source.Module] = moduleInstance
	}

	logger.mutex.Unlock()
	moduleInstance.mutex.Lock()

	clusterInstance, clusterFound := moduleInstance.clusters[source.Cluster]

	if !clusterFound {
		clusterInstance = NewCluster()
		moduleInstance.clusters[source.Cluster] = clusterInstance
	}

	moduleInstance.mutex.Unlock()
	clusterInstance.mutex.Lock()
	defer clusterInstance.mutex.Unlock()

	supervisorLogs, supervisorFound := clusterInstance.supervisors[source.Identifier]
	if !supervisorFound {
		supervisorLogs = make([]string, 0)
		clusterInstance.supervisors[source.Identifier] = supervisorLogs
	}

	l := record.(Log)
	supervisorLogs = append(supervisorLogs, l.ToString())

	return nil
}

func (logger *Logger) Flush(source message.Source, destination any) error {

	if !logger.enabled.logging {
		return message.LoggingDisabledError
	}

	logger.mutex.RLock()

	moduleInstance, moduleFound := logger.modules[source.Module]
	if !moduleFound {
		return message.ModuleNotFoundError
	}

	logger.mutex.RUnlock()
	moduleInstance.mutex.RLock()

	clusterInstance, clusterFound := moduleInstance.clusters[source.Cluster]
	if !clusterFound {
		return message.ClusterNotFoundError
	}

	moduleInstance.mutex.RUnlock()
	clusterInstance.mutex.RLock()
	defer clusterInstance.mutex.RUnlock()

	logs, logsFound := clusterInstance.supervisors[source.Identifier]

	if !logsFound {
		return message.RunnerNotFoundError
	}

	endpoint := fmt.Sprintf("%s_%s_%d", source.Module, source.Cluster, source.Identifier)

	if _, err := os.Stat(logger.directory); err != nil {
		return message.LogSaveFailedError
	}

	currTime := time.Now()
	currTimeStr := currTime.Format(time.RFC3339Nano)
	fileName := fmt.Sprintf("%s_%s.log", endpoint, currTimeStr)

	path := filepath.Join(logger.directory, fileName)
	cleanedPath := filepath.Clean(path)

	file, err := os.Create(cleanedPath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	for _, l := range logs {
		cleanedLog := strings.ReplaceAll(l, "\n", "")
		_, err = file.WriteString(cleanedLog + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
