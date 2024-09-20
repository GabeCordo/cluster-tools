package email

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/message"
	"sync"
)

var DefaultSmtpEndpoint = SmtpEndpoint{
	Host: "smtp.gmail.com",
	Port: "587",
}

type SmtpEndpoint struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
}

func (endpoint SmtpEndpoint) ToUrl() string {
	return fmt.Sprintf("%s:%s", endpoint.Host, endpoint.Port)
}

type SmtpCredentials struct {
	Email    string `yaml:"email" json:"email"`
	Password string `yaml:"password" json:"password"`
}

type SmtpRecord struct {
	Endpoint    SmtpEndpoint        `json:"endpoint" yaml:"endpoint"`
	Credentials SmtpCredentials     `json:"credentials" yaml:"credentials"`
	Receivers   map[string][]string `json:"receivers" yaml:"receivers"`
	Enabled     bool                `json:"enabled" yaml:"enabled"`
}

type Messenger struct {
	endpoint    SmtpEndpoint
	credentials SmtpCredentials

	receivers map[string][]string

	mutex sync.RWMutex
}

func New(endpoint SmtpEndpoint, credentials SmtpCredentials) *Messenger {

	msg := new(Messenger)

	msg.endpoint = endpoint
	msg.credentials = credentials

	return msg
}

func (messenger *Messenger) SetupSMTP(endpoint SmtpEndpoint, credentials SmtpCredentials) *Messenger {

	messenger.endpoint = endpoint
	messenger.credentials = credentials

	return messenger
}

func (messenger *Messenger) SetupReceivers(receivers map[string][]string) *Messenger {

	messenger.mutex.Lock()
	defer messenger.mutex.Unlock()

	messenger.receivers = receivers

	return messenger
}

func (messenger *Messenger) GetReceivers() map[string][]string {

	messenger.mutex.RLock()
	defer messenger.mutex.RUnlock()

	copyOfReceivers := make(map[string][]string)

	for endpoint, receivers := range messenger.receivers {
		copyOfReceivers[endpoint] = make([]string, len(receivers))
		copy(copyOfReceivers[endpoint], receivers)
	}

	return copyOfReceivers
}

func (messenger *Messenger) Message(source message.Source, record any) error {

	panic("not implemented")
}

func (messenger *Messenger) Flush(source message.Source, destination any) error {

	//emailSuccess := true
	//if messenger.enabled.smtp {
	//	if receivers, found := messenger.smtp.Receivers[endpoint]; found {
	//		emailSuccess = SendEmail(endpoint, messenger.smtp.Credentials, receivers, messenger.smtp.Endpoint)
	//	}
	//}
	panic("not implemented")
}
