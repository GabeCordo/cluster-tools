package email

import (
	"log"
	"net/smtp"
)

// SendEmail
// Returns a boolean representing whether an email was sent successfully
func sendEmail(message string, credentials SmtpCredentials, receivers []string, endpoints ...SmtpEndpoint) bool {
	endpoint := DefaultSmtpEndpoint
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
