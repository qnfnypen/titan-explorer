package config

var Cfg Config

type Config struct {
	EtcdAddress     string
	Mode            string
	ApiListen       string
	DatabaseURL     string
	SecretKey       string
	RedisAddr       string
	RedisPassword   string
	SchedulerFromDB bool
	IpKey           string
	IpUrl           string
	Locator         LocatorConfig
	Statistic       StatisticsConfig
	Email           EmailConfig
	AdminScheduler  AdminSchedulerConfig
}

type EmailConfig struct {
	Name     string
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

type LocatorConfig struct {
	Address       string
	Token         string
	AreaWhiteList []string
}

type StatisticsConfig struct {
	Disable bool
	Crontab string
}

type AdminSchedulerConfig struct {
	Enable  bool
	Address string
	Token   string
}
