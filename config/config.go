package config

var Cfg Config

var GNodesInfo NodesInfo

type NodesInfo struct {
	AssetCount int64
}

type Config struct {
	EtcdAddress              string
	Mode                     string
	ApiListen                string
	DatabaseURL              string
	SecretKey                string
	RedisAddr                string
	RedisPassword            string
	FilecoinRPCServerAddress string
	Statistic                StatisticsConfig
	Emails                   []EmailConfig
	StorageBackup            StorageBackupConfig
	IpDataCloud              IpDataCloudConfig
	ContainerManager         ContainerManagerEndpointConfig
	Epoch                    EpochConfig
	SpecifyCandidate         SpecifyCandidateConfig
	URL                      URLConfig
}

type EmailConfig struct {
	From     string
	Nickname string
	SMTPHost string
	SMTPPort string
	Username string
	Password string
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

type IpDataCloudConfig struct {
	Url string
	Key string
}

type ContainerManagerEndpointConfig struct {
	Addr  string
	Token string
}

type EpochConfig struct {
	Token string
}

type SpecifyCandidateConfig struct {
	Disable bool
	AreaId  string
	NodeId  string
}

type URLConfig struct {
	Discord string
}
