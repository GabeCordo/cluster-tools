package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/threads/cache"
	"github.com/GabeCordo/mango/core/threads/database"
	"github.com/GabeCordo/mango/core/threads/http_client"
	"github.com/GabeCordo/mango/core/threads/http_processor"
	"github.com/GabeCordo/mango/core/threads/messenger"
	"github.com/GabeCordo/mango/core/threads/processor"
	"github.com/GabeCordo/mango/core/threads/supervisor"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Config struct {
	Name               string  `yaml:"name"`
	Version            float64 `yaml:"version"`
	Debug              bool    `yaml:"debug"`
	HardTerminateTime  int     `yaml:"hard-terminate-time"`
	MaxWaitForResponse float64 `yaml:"max-wait-for-response"`
	MountByDefault     bool    `yaml:"mount-by-default"`
	EnableCors         bool    `yaml:"enable-cors"`
	EnableRepl         bool    `yaml:"enable-repl"`
	Cache              struct {
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
		} `yaml:"client"`
		Processor struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"processor"`
	} `yaml:"net"`
	Path string
}

// TODO : should this be here?
const (
	DefaultFilePermissions os.FileMode = 0755
)

func NewConfig(name string) *Config {
	config := new(Config)

	config.Name = name
	config.Version = 1.0

	config.EnableCors = false
	config.EnableRepl = false

	config.Net.Client.Port = 8136        // default
	config.Net.Client.Host = "localhost" // default

	config.Net.Processor.Port = 8137
	config.Net.Processor.Host = "localhost"

	return config
}

func (config *Config) Print() {

	bytes, _ := yaml.Marshal(config)
	fmt.Println(string(bytes))
}

func (config *Config) ToYAML(path string) {

	// if a core already exists, delete it
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	file, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
	}
	_ = ioutil.WriteFile(path, file, DefaultFilePermissions)
}

func (config *Config) Store() bool {
	// verify that the core file we initially loaded from has not been deleted
	if _, err := os.Stat(config.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}

	jsonRepOfConfig, err := json.Marshal(config)
	if err != nil {
		return false
	}

	err = os.WriteFile(config.Path, jsonRepOfConfig, 0666)
	if err != nil {
		return false
	}
	return true
}

func (config *Config) FillCacheConfig(clientConfig *cache.Config) {
	// TODO - add panic check
	clientConfig.Debug = config.Debug
}

func (config *Config) FillHttpClientConfig(httpClientConfig *http_client.Config) {
	// TODO - add panic check
	httpClientConfig.Debug = config.Debug
	httpClientConfig.Net.Host = config.Net.Client.Host
	httpClientConfig.Net.Port = config.Net.Client.Port
	httpClientConfig.EnableCors = config.EnableCors
	httpClientConfig.Timeout = config.MaxWaitForResponse
}

func (config *Config) FillHttpProcessorConfig(processorClientConfig *http_processor.Config) {
	// TODO - add panic check
	processorClientConfig.Debug = config.Debug
	processorClientConfig.Net.Host = config.Net.Processor.Host
	processorClientConfig.Net.Port = config.Net.Processor.Port
	processorClientConfig.Timeout = config.MaxWaitForResponse
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
}

func (config *Config) FillProcessorConfig(processorConfig *processor.Config) {
	// TODO - add panic check
	processorConfig.Debug = config.Debug
	processorConfig.Timeout = config.MaxWaitForResponse
}

func (config *Config) FillSupervisorConfig(supervisorConfig *supervisor.Config) {
	// TODO - add panic check
	supervisorConfig.Debug = config.Debug
	supervisorConfig.Timeout = config.MaxWaitForResponse
}

func YAMLToETLConfig(config *Config, path string) error {
	if _, err := os.Stat(path); err != nil {
		// file does not exist
		log.Println(err)
		return err
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		// error reading the file
		log.Println(err)
		return err
	}

	err = yaml.Unmarshal([]byte(file), config)
	if err != nil {
		// the file is not a JSON or is a malformed (fields missing) core
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
			ConfigInstance.Path = configPath[0]
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
