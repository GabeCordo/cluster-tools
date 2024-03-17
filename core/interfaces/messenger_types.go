package interfaces

type SmtpEndpoint struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
}

type SmtpCredentials struct {
	Email    string `yaml:"email" json:"email"`
	Password string `yaml:"password" json:"password"`
}
