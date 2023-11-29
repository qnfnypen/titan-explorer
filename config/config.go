package config

var Cfg Config

var GNodesInfo NodesInfo

type NodesInfo struct {
	AssetCount int64
}

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
	StorageBackup   StorageBackupConfig
}

type EmailConfig struct {
	From     string
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

type StorageBackupConfig struct {
	BackupPath string
	Crontab    string
	Disable    bool
}
