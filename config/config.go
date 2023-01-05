package config

type Config struct {
	Mode          string
	ApiListen     string
	DatabaseURL   string
	SecretKey     string
	RedisAddr     string
	RedisPassword string
	Locator       LocatorConfig
	Email         EmailConfig
}

type EmailConfig struct {
	SMTP    string
	Host    string
	Name    string
	Address string
	Secret  string
}

type LocatorConfig struct {
	Address   string
	Token     string
	Enable    bool
	WhiteList []string
}
