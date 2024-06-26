package messenger

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
)

func (messenger *Messenger) LoggingDirectory(path string) *Messenger {
	if messenger.enabled.logging {
		messenger.logging.directory = path
	}

	return messenger
}

func (messenger *Messenger) SetupSMTP(endpoint interfaces.SmtpEndpoint, credentials interfaces.SmtpCredentials) *Messenger {
	if messenger.enabled.smtp {
		messenger.smtp.Endpoint = endpoint
		messenger.smtp.Credentials = credentials
	}

	return messenger
}

func (messenger *Messenger) SetupReceivers(receivers map[string][]string) *Messenger {

	messenger.mutex.Lock()
	defer messenger.mutex.Unlock()

	if messenger.enabled.smtp {
		messenger.smtp.Receivers = receivers
	}

	return messenger
}

func (messenger *Messenger) GetReceivers() map[string][]string {

	messenger.mutex.RLock()
	defer messenger.mutex.RUnlock()

	copyOfReceivers := make(map[string][]string)

	for endpoint, receivers := range messenger.smtp.Receivers {
		copyOfReceivers[endpoint] = make([]string, len(receivers))
		copy(copyOfReceivers[endpoint], receivers)
	}

	return copyOfReceivers
}

func (messenger *Messenger) Get(identifier string) (*Module, bool) {

	messenger.mutex.RLock()
	defer messenger.mutex.RUnlock()

	instance, found := messenger.modules[identifier]
	return instance, found
}

func (messenger *Messenger) Create(identifier string) (*Module, error) {

	messenger.mutex.Lock()
	defer messenger.mutex.Unlock()

	if _, found := messenger.modules[identifier]; found {
		return nil, errors.New("module already exists")
	}

	module := NewModule()
	messenger.modules[identifier] = module
	return module, nil
}

// Log
// a facade for generating logs for a given supervisor
func (messenger *Messenger) Log(module, cluster string, supervisor uint64, message string, priority ...MessagePriority) error {

	moduleInstance, moduleFound := messenger.Get(module)

	if !moduleFound {
		moduleInstance, _ = messenger.Create(module)
	}

	clusterInstance, clusterFound := moduleInstance.Get(cluster)

	if !clusterFound {
		clusterInstance, _ = moduleInstance.Create(cluster)
	}

	level := Normal
	for _, p := range priority {
		level = p
	}

	return clusterInstance.Add(supervisor, level, message)
}

// Complete
// facade for closing the log and generating the log file for a supervisor
func (messenger *Messenger) Complete(module, cluster string, supervisor uint64) (success bool) {

	success = false
	moduleInstance, moduleFound := messenger.Get(module)

	if !moduleFound {
		return success
	}

	clusterInstance, clusterFound := moduleInstance.Get(cluster)

	if !clusterFound {
		return success
	}

	logs, logsFound := clusterInstance.Get(supervisor)

	if !logsFound {
		return success
	}

	endpoint := fmt.Sprintf("%s_%s_%d", module, cluster, supervisor)

	emailSuccess := true
	if messenger.enabled.smtp {
		if receivers, found := messenger.smtp.Receivers[endpoint]; found {
			emailSuccess = SendEmail(endpoint, messenger.smtp.Credentials, receivers, messenger.smtp.Endpoint)
		}
	}

	loggingSuccess := true
	if messenger.enabled.logging {
		loggingSuccess = SaveToFile(messenger.logging.directory, endpoint, logs)
	}

	success = emailSuccess && loggingSuccess
	return success
}
