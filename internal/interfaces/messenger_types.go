package interfaces

import "fmt"

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
