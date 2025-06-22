package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/thread/cache"
	"github.com/GabeCordo/Flock/internal/core/thread/database"
	"github.com/GabeCordo/Flock/internal/core/thread/messenger"
	"github.com/GabeCordo/Flock/internal/core/thread/processor"
	httpClient "github.com/GabeCordo/Flock/internal/core/thread/rest"
	"github.com/GabeCordo/Flock/internal/core/thread/runner"
	"github.com/GabeCordo/Flock/internal/core/thread/scheduler"
	"github.com/GabeCordo/Flock/internal/core/thread/socket"
	"gopkg.in/yaml.v3"
)

const defaultFilePerm = 0600

type Config struct {
	Name               string  `yaml:"name"`
	Version            float64 `yaml:"version"`
	Debug              bool    `yaml:"debug"`
	HardTerminateTime  int     `yaml:"hard-terminate-time"`
	MaxWaitForResponse float64 `yaml:"max-wait-for-response"`
	MountByDefault     bool    `yaml:"mount-by-default"`
	EnableCors         bool    `yaml:"enable-cors"`
	EnableRepl         bool    `yaml:"enable-repl"`
	Database           struct {
		Type string `yaml:"type"`
	} `yaml:"database"`
	Cache struct {
		Expiry  float64 `yaml:"expire-in"`
		MaxSize uint32  `yaml:"max-size"`
	} `yaml:"cache"`
	Messenger struct {
		LogFiles struct {
			Directory string `yaml:"directory"`
		} `yaml:"logging,omitempty"`
		EnableLogging bool `yaml:"enable-logging"`
		Smtp          struct {
			Endpoint struct {
				Host string `yaml:"host"`
				Port string `yaml:"port"`
			} `yaml:"endpoint"`
			Credentials struct {
				Email    string `yaml:"email"`
				Password string `yaml:"password"`
			} `yaml:"credentials"`
			Subscribers map[string][]string `yaml:"subscribers"`
		} `json:"smtp,omitempty"`
		EnableSmtp bool `yaml:"enable-smtp"`
	} `json:"messenger"`
	Net struct {
		Client struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"rest"`
		Processor struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
			TLS  struct {
				Certificate string `yaml:"certificate"`
				PrivateKey  string `yaml:"private_key"`
			} `yaml:"TLS"`
		} `yaml:"processor"`
	} `yaml:"net"`
	Processor struct {
		ProbeEvery uint32 `yaml:"probe-every"`
		MaxRetry   uint32 `yaml:"max-retry"`
	} `yaml:"processor"`
	Paths struct {
		Root       string `yaml:"root"`
		Configs    string `yaml:"configs"`
		Logs       string `yaml:"logs"`
		Statistics string `yaml:"statistics"`
		Schedules  string `yaml:"schedules"`
		Messenger  string `yaml:"messenger"`
	} `yaml:"paths"`
}

func NewConfig(name string) *Config {
	config := new(Config)

	config.Name = name
	config.Version = 1.0

	config.EnableCors = false
	config.EnableRepl = false

	config.Database.Type = "file"

	config.MaxWaitForResponse = 2
	config.MountByDefault = true

	config.Processor.ProbeEvery = 10
	config.Processor.MaxRetry = 5

	config.Net.Client.Port = 8136        // default
	config.Net.Client.Host = "localhost" // default

	config.Net.Processor.Port = 8137
	config.Net.Processor.Host = "localhost"

	config.Processor.ProbeEvery = 2
	config.Processor.MaxRetry = 10

	return config
}

func (config *Config) Print() {

	bytes, _ := yaml.Marshal(config)
	fmt.Println(string(bytes))
}

func (config *Config) ToYAML(path string) error {

	// if a flock already exists, delete it
	_, err := os.Stat(path)
	if err == nil {
		// attempt to remove the path
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	file, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	path = filepath.Clean(path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = f.Write(file) // todo: do we need DefaultFilePermissions?
	return err
}

func (config *Config) Store() bool {
	// verify that the flock file we initially loaded from has not been deleted
	if _, err := os.Stat(config.Paths.Root); errors.Is(err, os.ErrNotExist) {
		return false
	}

	jsonRepOfConfig, err := json.Marshal(config)
	if err != nil {
		return false
	}

	err = os.WriteFile(config.Paths.Root, jsonRepOfConfig, defaultFilePerm)
	if err != nil {
		return false
	}
	return true
}

func (config *Config) FillCacheConfig(clientConfig *cache.Config) {
	// TODO - add panic check
	clientConfig.Debug = config.Debug
}

func (config *Config) FillHttpClientConfig(httpClientConfig *httpClient.Config) {
	// TODO - add panic check
	httpClientConfig.Debug = config.Debug
	httpClientConfig.Net.Host = config.Net.Client.Host
	httpClientConfig.Net.Port = config.Net.Client.Port
	httpClientConfig.EnableCors = config.EnableCors
	httpClientConfig.Timeout = config.MaxWaitForResponse
}

func (config *Config) FillSocketConfig(socketConfig *socket.Config) {
	// TODO - add panic check
	socketConfig.Debug = config.Debug
	socketConfig.Net.Host = config.Net.Processor.Host
	socketConfig.Net.Port = config.Net.Processor.Port
	socketConfig.Timeout = config.MaxWaitForResponse
	socketConfig.Tls.Certificate = config.Net.Processor.TLS.Certificate
	socketConfig.Tls.Key = config.Net.Processor.TLS.PrivateKey
}

func (config *Config) FillMessengerConfig(messengerConfig *messenger.Config) {
	// TODO - add panic check
	messengerConfig.Debug = config.Debug
	messengerConfig.EnableLogging = config.Messenger.EnableLogging
	messengerConfig.LoggingDir = config.Messenger.LogFiles.Directory
	messengerConfig.EnableSmtp = config.Messenger.EnableSmtp
	messengerConfig.SmtpEndpoint.Host = config.Messenger.Smtp.Endpoint.Host
	messengerConfig.SmtpEndpoint.Port = config.Messenger.Smtp.Endpoint.Port
	messengerConfig.SmtpCredentials.Email = config.Messenger.Smtp.Credentials.Email
	messengerConfig.SmtpCredentials.Password = config.Messenger.Smtp.Credentials.Password
	messengerConfig.SmtpSubscribers = config.Messenger.Smtp.Subscribers
}

func (config *Config) FillDatabaseConfig(databaseConfig *database.Config) {
	databaseConfig.Debug = config.Debug
	databaseConfig.Timeout = config.MaxWaitForResponse
	databaseConfig.Type = config.Database.Type
	databaseConfig.ConfigsFolder = config.Paths.Configs
	databaseConfig.StatisticsFolder = config.Paths.Statistics
}

func (config *Config) FillProcessorConfig(processorConfig *processor.Config) {
	// TODO - add panic check
	processorConfig.Debug = config.Debug
	processorConfig.Timeout = config.MaxWaitForResponse
	processorConfig.MaxRetry = config.Processor.MaxRetry
	processorConfig.ProbeEvery = config.Processor.ProbeEvery
	processorConfig.Net.Host = config.Net.Processor.Host
	processorConfig.Net.Port = config.Net.Processor.Port
}

func (config *Config) FillRunnerConfig(supervisorConfig *runner.Config) {
	// TODO - add panic check
	supervisorConfig.Debug = config.Debug
	supervisorConfig.Timeout = config.MaxWaitForResponse
}

func (config *Config) FillSchedulerConfig(schedulerConfig *scheduler.Config) {
	schedulerConfig.Debug = config.Debug
	schedulerConfig.Timeout = config.MaxWaitForResponse
	schedulerConfig.SchedulesFolder = config.Paths.Schedules
}

func YAMLToETLConfig(config *Config, path string) error {

	if _, err := os.Stat(path); err != nil {
		// file does not exist
		return err
	}

	path = filepath.Clean(path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	err = yaml.NewDecoder(f).Decode(&config)
	if err != nil {
		// the file is not a JSON or is a malformed (fields missing) flock
		log.Println(err)
		return err
	}

	return nil
}

var (
	configLock     = &sync.Mutex{}
	ConfigInstance *Config
)

func GetConfigInstance(configPath ...string) *Config {
	configLock.Lock()
	defer configLock.Unlock()

	/* if this is the first time the common is being loaded the develoepr
	   needs to pass in a configPath to load the common instance from
	*/
	if (ConfigInstance == nil) && (len(configPath) < 1) {
		return nil
	}

	if ConfigInstance == nil {
		ConfigInstance = NewConfig("test")

		if err := YAMLToETLConfig(ConfigInstance, configPath[0]); err == nil {
			// the configPath we found the common for future reference
			ConfigInstance.Paths.Root = configPath[0]
			// if the Timeout is not set, then simply default to 2.0
			if ConfigInstance.MaxWaitForResponse == 0 {
				ConfigInstance.MaxWaitForResponse = 2
			}
		} else {
			log.Println("(!) the etl configuration file can either not be found or is corrupted")
			log.Fatal(fmt.Sprintf("%s was not a valid common path\n", configPath))
		}
	}

	return ConfigInstance
}
