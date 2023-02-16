package config

type Config struct {
	Mode          string
	ApiListen     string
	DatabaseURL   string
	SecretKey     string
	RedisAddr     string
	RedisPassword string
	Locator       LocatorConfig
	Statistic     StatisticsConfig
	Email         EmailConfig
	Admin         AdminConfig
}

type EmailConfig struct {
	SMTP    string
	Host    string
	Name    string
	Address string
	Secret  string
}

type LocatorConfig struct {
	Address       string
	Token         string
	Enable        bool
	AreaWhiteList []string
}

type StatisticsConfig struct {
	Disable bool
	Crontab string
}

type AdminConfig struct {
	SchedulerURL string
	Token        string
}
