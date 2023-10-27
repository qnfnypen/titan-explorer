package model

import "time"

type Language string

const (
	LanguageEN = "en"
	LanguageCN = "cn"
)

type Project struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type StorageStats struct {
	ID                         int64     `db:"id" json:"id"`
	Rank                       int64     `db:"rank" json:"rank"`
	ProjectId                  int64     `db:"project_id" json:"project_id"`
	ProjectName                string    `db:"project_name" json:"project_name"`
	TotalSize                  int64     `db:"total_size" json:"total_size"`
	UserCount                  int64     `db:"user_count" json:"user_count"`
	ProviderCount              int64     `db:"provider_count" json:"provider_count"`
	Expiration                 string    `db:"expiration" json:"expiration"`
	Time                       string    `db:"time" json:"time"`
	StorageChange24H           int64     `db:"storage_change_24h" json:"storage_change_24h"`
	StorageChangePercentage24H float64   `db:"storage_change_percentage_24h" json:"storage_change_percentage_24h"`
	Gas                        float64   `db:"gas" json:"gas"`
	Pledge                     float64   `db:"pledge" json:"pledge"`
	CreatedAt                  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                  time.Time `db:"updated_at" json:"updated_at"`
	ProviderIds                string    `db:"provider_ids" json:"-"`
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
}

type StorageProvider struct {
	ID              int64     `db:"id" json:"id"`
	ProviderID      string    `db:"provider_id" json:"provider_id"`
	IP              string    `db:"ip" json:"ip"`
	Location        string    `db:"location" json:"location"`
	Retrievable     string    `db:"retrievable" json:"retrievable"`
	NationalFlagUrl string    `db:"national_flag_url" json:"national_flag_url"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
