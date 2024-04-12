package messenger

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"log"
	"net/smtp"
)

// SendEmail
// Returns a boolean representing whether an email was sent successfully
func SendEmail(message string, credentials interfaces.SmtpCredentials, receivers []string, endpoints ...interfaces.SmtpEndpoint) bool {
	endpoint := interfaces.DefaultSmtpEndpoint
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
