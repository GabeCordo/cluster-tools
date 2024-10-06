package log

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/toolchain/files"
	"github.com/Sentmint/cluster-tools/internal/core/message"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var logRegex = regexp.MustCompile(`\[(.+)\]\[(.+)\](.+)`)

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

	logger.mutex.RLock()

	moduleInstance, moduleFound := logger.modules[source.Module]
	if !moduleFound {
		return errors.New("module not found")
	}

	logger.mutex.RUnlock()
	moduleInstance.mutex.RLock()

	clusterInstance, clusterFound := moduleInstance.clusters[source.Cluster]
	if !clusterFound {
		return errors.New("cluster not found")
	}

	moduleInstance.mutex.RUnlock()
	clusterInstance.mutex.RLock()
	defer clusterInstance.mutex.RUnlock()

	logs, logsFound := clusterInstance.supervisors[source.Identifier]

	if !logsFound {
		return errors.New("runner not found")
	}

	endpoint := fmt.Sprintf("%s_%s_%d", source.Module, source.Cluster, source.Identifier)

	if !logger.enabled.logging {
		return errors.New("cannot log when logging is temp. disabled")
	}

	if _, err := os.Stat(logger.directory); err != nil {
		return errors.New("warning: cannot save logs to file, the save directory doesn't exist")
	}

	currTime := time.Now()
	currTimeStr := currTime.Format(time.RFC3339Nano)
	fileName := fmt.Sprintf("%s_%s.log", endpoint, currTimeStr)

	path := files.EmptyPath().Dir(logger.directory).File(fileName)

	file, err := path.Create()
	if err != nil {
		return err
	}
	defer file.Close()

	for _, l := range logs {
		cleanedLog := strings.ReplaceAll(l, "\n", "")
		file.WriteString(cleanedLog + "\n")
	}

	return nil
}
