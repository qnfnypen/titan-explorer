package config

var Cfg Config

type Config struct {
	Mode                     string
	ApiListen                string
	DatabaseURL              string
	QuestDatabaseURL         string
	PlatformDatabaseURL      string
	SecretKey                string
	RedisAddr                string
	RedisPassword            string
	FilecoinRPCServerAddress string
	EtcdAddresses            []string
	EligibleOnlineMinutes    int
	ResourcePath             string
	Statistic                StatisticsConfig
	Emails                   []EmailConfig
	IpDataCloud              IpDataCloudConfig
	Epoch                    EpochConfig
	SpecifyCandidate         SpecifyCandidateConfig
	URL                      URLConfig
	Oss                      OssConfig
	Locators                 []string
	BaseURL                  string

	KubesphereAPI KubesphereAPIConfig
	ChainAPI      ChainAPIConfig
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

type IpDataCloudConfig struct {
	Url string
	Key string
}

type EpochConfig struct {
	Token string
}

type SpecifyCandidateConfig struct {
	Disable bool
	AreaId  string
}

type URLConfig struct {
	Discord string
}

type OssConfig struct {
	EndPoint  string
	AccessId  string
	AccessKey string
	Bucket    string
	Host      string
}

// KubesphereAPIConfig holds the configuration for the KubeSphere API.
type KubesphereAPIConfig struct {
	URL           string
	AdminAccount  string
	AdminPassword string
	Cluster       string
}

// ChainAPIConfig holds the configuration for the chain API.
type ChainAPIConfig struct {
	AddressPrefix        string
	RPC                  string
	TokenContractAddress string
	ServiceName          string
	KeyringDir           string
	FaucetGas            string
	OrderContractAddress string
}
