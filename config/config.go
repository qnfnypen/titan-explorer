package config

type Config struct {
	Mode          string
	ApiListen     string
	DatabaseURL   string
	SecretKey     string
	RedisAddr     string
	RedisPassword string
	Email         EmailConfig
}

type EmailConfig struct {
	SMTP    string
	Host    string
	Name    string
	Address string
	Secret  string
}
