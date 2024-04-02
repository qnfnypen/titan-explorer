package model

import "time"

type Language string

const (
	LanguageEN = "en"
	LanguageCN = "cn"
)

var SupportLanguages = []Language{LanguageEN, LanguageCN}

type RewardEvent string

var (
	RewardEventInviteFrens RewardEvent = "invite_frens"
	RewardEventBindDevice  RewardEvent = "bind_device"
	RewardEventEarning     RewardEvent = "earning"
	RewardEventReferrals   RewardEvent = "referrals"
)

type Project struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

type StorageStats struct {
	ID                         int64     `db:"id" json:"id"`
	Rank                       int64     `db:"s_rank" json:"s_rank"`
	ProjectId                  int64     `db:"project_id" json:"project_id"`
	ProjectName                string    `db:"project_name" json:"project_name"`
	TotalSize                  int64     `db:"total_size" json:"total_size"`
	UserCount                  int64     `db:"user_count" json:"user_count"`
	ProviderCount              int64     `db:"provider_count" json:"provider_count"`
	Expiration                 time.Time `db:"expiration" json:"expiration"`
	Time                       string    `db:"time" json:"time"`
	StorageChange24H           int64     `db:"storage_change_24h" json:"storage_change_24h"`
	StorageChangePercentage24H float64   `db:"storage_change_percentage_24h" json:"storage_change_percentage_24h"`
	Gas                        float64   `db:"gas" json:"gas"`
	Pledge                     float64   `db:"pledge" json:"pledge"`
	Locations                  string    `db:"locations" json:"locations"`
	CreatedAt                  time.Time `db:"created_at" json:"-"`
	UpdatedAt                  time.Time `db:"updated_at" json:"-"`
}

type StorageSummary struct {
	TotalSize         float64 `db:"total_size" json:"total_size"`
	Projects          int64   `db:"projects" json:"projects"`
	Users             int64   `db:"users" json:"users"`
	Pledges           float64 `db:"pledges" json:"pledges"`
	Gases             float64 `db:"gases" json:"gases"`
	Providers         int64   `db:"providers" json:"providers"`
	RetrievalProvider int64   `db:"retrieval_providers" json:"retrieval_providers"`
	StorageProvider   int64   `db:"storage_providers" json:"storage_providers"`
	LatestUpdateTime  string  `db:"-" json:"latest_update_time"`
}

type StorageProvider struct {
	ID          int64     `db:"id" json:"id"`
	ProviderID  string    `db:"provider_id" json:"provider_id"`
	IP          string    `db:"ip" json:"ip"`
	Location    string    `db:"location" json:"location"`
	Retrievable bool      `db:"retrievable" json:"retrievable"`
	CreatedAt   time.Time `db:"created_at" json:"-"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}

type InviteFrensRecord struct {
	Email      string    `db:"email" json:"email"`
	Status     int       `db:"status" json:"status"`
	BoundCount int       `db:"bound_count" json:"bound_count"`
	Reward     float64   `db:"reward" json:"reward"`
	Time       time.Time `db:"time" json:"time"`
}

type SignInfo struct {
	MinerID      string `json:"miner_id" db:"miner_id"`
	Address      string `json:"address" db:"address"`
	Date         int64  `json:"date" db:"date"`
	SignedMsg    string `json:"signed_msg" db:"signed_msg"`
	MinerPower   string `json:"miner_power" db:"miner_power"`
	MinerBalance string `json:"miner_balance" db:"miner_balance"`
}

type DeviceDistribution struct {
	Country string `json:"country" db:"country"`
	Count   int    `json:"count" db:"count"`
}

type AppVersion struct {
	ID          int64     `db:"id" json:"-"`
	Version     string    `db:"version" json:"version"`
	MinVersion  string    `db:"min_version" json:"min_version"`
	Description string    `db:"description" json:"description"`
	Url         string    `db:"url" json:"url"`
	Size        int64     `db:"size" json:"size"`
	Platform    string    `db:"platform" json:"platform"`
	Lang        string    `db:"lang" json:"lang"`
	CreatedAt   time.Time `db:"created_at" json:"-"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}
