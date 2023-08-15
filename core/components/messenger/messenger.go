package messenger

import (
	"fmt"
	"sync"
	"time"
)

type MessagePriority uint

const (
	Log MessagePriority = iota
	Warning
	Fatal
)

func (priority MessagePriority) ToString() string {
	if priority == Log {
		return "-"
	} else if priority == Warning {
		return "?"
	} else {
		return "!"
	}
}

type Messenger struct {
	enabled struct {
		logging bool
		smtp    bool
	}
	logging struct {
		directory string
	}
	smtp struct {
		endpoint    Endpoint
		credentials Credentials
		receivers   map[string][]string
	}

	data  map[string][]string
	mutex sync.Mutex
}

func NewMessenger(enableLogging, enableSmtp bool) *Messenger {
	messenger := new(Messenger)
	messenger.data = make(map[string][]string)

	messenger.enabled.logging = enableLogging
	messenger.enabled.smtp = enableSmtp

	return messenger
}

func (messenger *Messenger) LoggingDirectory(path string) *Messenger {
	if messenger.enabled.logging {
		messenger.logging.directory = path
	}

	return messenger
}

func (messenger *Messenger) SetupSMTP(endpoint Endpoint, credentials Credentials) *Messenger {
	if messenger.enabled.smtp {
		messenger.smtp.endpoint = endpoint
		messenger.smtp.credentials = credentials
	}

	return messenger
}

func (messenger *Messenger) SetupReceivers(receivers map[string][]string) *Messenger {
	if messenger.enabled.smtp {
		messenger.smtp.receivers = receivers
	}

	return messenger
}

func (messenger *Messenger) Log(endpoint, message string, priority ...MessagePriority) {

	messenger.mutex.Lock()
	defer messenger.mutex.Unlock()

	priorityStr := Log.ToString()
	if len(priority) != 0 {
		priorityStr = priority[0].ToString()
	}

	log := fmt.Sprintf("[%s][%s] %s", time.Now().Format("2006-01-02 15:04:05"), priorityStr, message)

	if logs, found := messenger.data[endpoint]; found {
		messenger.data[endpoint] = append(logs, log)
	} else {
		logs := make([]string, 0)
		logs = append(logs, log)
		messenger.data[endpoint] = logs
	}
}

func (messenger *Messenger) Warning(endpoint, message string) {

}

func (messenger *Messenger) Complete(endpoint string) bool {

	messenger.mutex.Lock()
	defer messenger.mutex.Unlock()

	var success bool = false

	if logs, found := messenger.data[endpoint]; !found {
		success = false
	} else {
		message := fmt.Sprintf("Cluster: %s\n", endpoint)

		for _, log := range logs {
			message += fmt.Sprintf("\n%s", log)
		}

		emailSuccess := true
		if messenger.enabled.smtp {
			if receivers, found := messenger.smtp.receivers[endpoint]; found {
				emailSuccess = SendEmail(message, messenger.smtp.credentials, receivers, messenger.smtp.endpoint)
			}
		}

		loggingSuccess := true
		if messenger.enabled.logging {
			loggingSuccess = SaveToFile(messenger.logging.directory, endpoint, logs)
		}

		success = emailSuccess && loggingSuccess
	}

	return success
}
