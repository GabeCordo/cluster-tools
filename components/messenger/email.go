package messenger

import (
	"fmt"
	"log"
	"net/smtp"
)

type Credentials struct {
	Email    string
	Password string
}

type Endpoint struct {
	Host string
	Port string
}

func (endpoint Endpoint) ToUrl() string {
	return fmt.Sprintf("%s:%s", endpoint.Host, endpoint.Port)
}

var DefaultEndpoint Endpoint = Endpoint{
	Host: "smtp.gmail.com",
	Port: "587",
}

// SendEmail
// Returns a boolean representing whether an email was sent successfully
func SendEmail(message string, credentials Credentials, receivers []string, endpoints ...Endpoint) bool {
	endpoint := DefaultEndpoint
	if len(endpoints) == 1 {
		endpoint = endpoints[0]
	}

	bytes := []byte(message)
	auth := smtp.PlainAuth("", credentials.Email, credentials.Password, endpoint.Host)

	err := smtp.SendMail(endpoint.ToUrl(), auth, credentials.Email, receivers, bytes)
	if err != nil {
		log.Println(err)
	}

	return err == nil
}
