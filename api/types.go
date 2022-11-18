package api

import (
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

// IndexPageRes structure of index info
type IndexPageRes struct {
	// AllMinerNum MinerInfo
	AllMinerInfo
	// OnlineMinerNum MinerInfo
	OnlineVerifier  int `json:"online_verifier"`  // 在线验证人
	OnlineCandidate int `json:"online_candidate"` // 在线候选人
	OnlineEdgeNode  int `json:"online_edge_node"` // 在线边缘节点
	// ProfitInfo Profit  // 个人收益信息
	ProfitInfo
	// Device Devices // 设备信息
	MinerDevices
}

type IndexUserDeviceRes struct {
	IndexUserDevice
	DailyIncome interface{} `json:"daily_income"` // 日常收益
}

// IndexUserDevice 个人设备总览
type IndexUserDevice struct {
	// ProfitInfo Profit  // 个人收益信息
	ProfitInfo
	// Device Devices // 设备信息
	MinerDevices
}

// MinerDevices Device Devices // 设备信息
type MinerDevices struct {
	TotalNum       int64   `json:"total_num"`       // 设备总数
	OnlineNum      int64   `json:"online_num"`      // 在线设备数
	OfflineNum     int64   `json:"offline_num"`     // 离线设备数
	AbnormalNum    int64   `json:"abnormal_num"`    // 异常设备数
	TotalBandwidth float64 `json:"total_bandwidth"` // 总上行速度（kbps）
}
type ProfitInfo struct {
	CumulativeProfit float64 `json:"cumulative_profit"` // 个人累计收益
	YesterdayProfit  float64 `json:"yesterday_profit"`  // 昨日收益
	TodayProfit      float64 `json:"today_profit"`      // 今日收益
	SevenDaysProfit  float64 `json:"seven_days_profit"` // 近七天收益
	MonthProfit      float64 `json:"month_profit"`      // 近30天收益
}

type AllMinerInfo struct {
	AllVerifier  int     `json:"all_verifier"`  // 全网验证人
	AllCandidate int     `json:"all_candidate"` // 全网候选人
	AllEdgeNode  int     `json:"all_edgeNode"`  // 全网边缘节点
	StorageT     float64 `json:"storage_t"`     // 全网存储（T）
	BandwidthMb  float64 `json:"bandwidth_mb"`  // 全网上行带宽（MB/S）
}

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     string `json:"page" form:"page"`         // 页码
	PageSize string `json:"pageSize" form:"pageSize"` // 每页大小
	Data     string `json:"data" form:"data"`         // 关键字
	DateFrom string `json:"dateFrom" form:"dateFrom"` // 日期开始
	DateTo   string `json:"dateTo" form:"dateTo"`     // 日期结束
	Date     string `json:"date" form:"date"`         // 具体日期
	Device   string `json:"deviceId" form:"deviceId"` // 设备ID
	UserIds  string `json:"userId" form:"userId"`     // 用户ID
	UserIp   string `json:"userIp" form:"userIp"`     // user ip address
}

// RetrievalInfo  miner info
type RetrievalInfo struct {
	ServiceCountry string  `json:"service_country"` // 服务商国家
	ServiceStatus  string  `json:"service_status"`  // 服务商网络状态
	TaskStatus     string  `json:"task_status"`     // 任务状态
	FileName       string  `json:"file_name"`       // 文件名
	FileSize       string  `json:"file_size"`       // 文件大小
	CreateTime     string  `json:"create_time"`     // 文件创建日期
	Cid            string  `json:"cid"`             // 编号
	Price          float64 `json:"price"`           // 价格
	MinerId        string  `json:"miner_id"`        // 矿工id
	UserId         string  `json:"user_id"`         // 用户id
	DownloadUrl    string  `json:"download_url"`    // 下载地址
}

// RetrievalPageRes Response data of Retrieval miner info
type RetrievalPageRes struct {
	List []*model.TaskInfo `json:"list"`
	AllMinerInfo
	Count int64 `json:"count"`
}

type Base struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"id" gorm:"comment:'创建时间';type:timestamp;"`
	UpdatedAt time.Time `json:"id" gorm:"comment:'更新时间';type:timestamp;"`
}

type RpcDevice struct {
	JsonRpc string           `json:"jsonrpc"`
	Id      int              `json:"id"`
	Result  model.DeviceInfo `json:"result"`
}

type RpcTask struct {
	JsonRpc string            `json:"jsonrpc"`
	Id      int               `json:"id"`
	Result  []TaskDataFromRpc `json:"result"`
}

type DeviceInfoPage struct {
	List  []*model.DeviceInfo `json:"list"`
	Count int64               `json:"count"`
	DeviceType
}

type DeviceType struct {
	Online      int64   `json:"online"`       // 在线
	Offline     int64   `json:"offline"`      // 离线
	Abnormal    int64   `json:"abnormal"`     // 异常
	AllDevices  int64   `json:"all_devices"`  // 全部设备
	BandwidthMb float64 `json:"bandwidth_mb"` // 全网上行带宽（MB/S）
}

type IncomeDailySearch struct {
	PageInfo
	model.IncomeDaily
}

type IncomeDailyRes struct {
	DailyIncome      interface{} `json:"daily_income"`      // 日常收益
	DefYesterday     string      `json:"def_yesterday"`     // 较昨日
	CumulativeProfit float64     `json:"cumulative_profit"` // 累计收益
	YesterdayProfit  float64     `json:"yesterday_profit"`  // 昨日收益
	SevenDaysProfit  float64     `json:"seven_days_profit"` // 近七天
	MonthProfit      float64     `json:"month_profit"`      // 近30天
	TodayProfit      float64     `json:"today_profit"`      // 今天收益
	OnlineTime       string      `json:"online_time"`       // 在线时长
	HighOnlineRatio  string      `json:"high_online_ratio"` // 高峰期在线率
	DeviceDiagnosis  string      `json:"diagnosis"`         // 诊断
}

type TaskDataFromRpc struct {
	Base
	// 用户id
	UserId string `json:"userId" form:"userId" gorm:"column:user_id;comment:;"`
	// 矿工id
	MinerId string `json:"minerId" form:"minerId" gorm:"column:miner_id;comment:;"`
	// 设备id
	DeviceId string `json:"deviceId" form:"deviceId" gorm:"column:device_id;comment:;"`
	// 请求cid
	Cid string `json:"blockCid" form:"cid" gorm:"column:cid;comment:;"`
	// 目的地址
	IpAddress string `json:"ipAddress" form:"cid" gorm:"column:ip_address;comment:;"`
	// 请求cid
	FileSize float64 `json:"blockSize" form:"fileSize" gorm:"column:file_size;comment:;"`
	// 文件名
	FileName float64 `json:"fileName" form:"fileName" gorm:"column:file_name;comment:;"`
	// 上行带宽B/s
	BandwidthUp float64 `json:"speed" form:"speed" gorm:"column:bandwidth_up;comment:;"`
	// 下行带宽B/s
	BandwidthDown float64 `json:"bandwidth_down" form:"bandwidth_down" gorm:"column:bandwidth_down;comment:;"`
	// 期望完成时间
	TimeNeed string `json:"time_need" form:"timeNeed" gorm:"column:time_need;comment:;"`
	// 完成时间
	TimeDone time.Time `json:"createdAt" form:"time" gorm:"column:time;comment:;"`
	// 服务商国家
	ServiceCountry string `json:"serviceCountry" form:"serviceCountry" gorm:"column:service_country;comment:;"`
	// 地区
	Region string `json:"region" form:"region" gorm:"column:region;comment:;"`
	// 当前状态
	Status string `json:"status" form:"status" gorm:"column:status;comment:;"`
	// 价格
	Reward float64 `json:"reward" form:"reward" gorm:"column:price;comment:;"`
	// 下载地址
	DownloadUrl float64 `json:"downloadUrl" form:"downloadUrl" gorm:"column:download_url;comment:;"`
}
