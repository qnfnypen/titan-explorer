package main

import (
	"context"
	"log"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx/v3"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("reading config file: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unmarshaling config file: %v\n", err)
	}

	if err := dao.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	ctx := context.Background()

	users, err := dao.AllUsers(ctx)
	if err != nil {
		log.Fatalf("fetching users: %v\n", err)
	}
	var em = make(map[string]*model.User)
	for _, u := range users {
		em[u.Username] = u
	}

	assets, err := dao.AllAssets(ctx)
	if err != nil {
		log.Fatalf("fetching assets: %v\n", err)
	}
	var ea = make(map[string]*model.Asset)
	for _, a := range assets {
		ea[a.Hash] = a
	}

	eusers, err := MergeUserInfo(em)
	if err != nil {
		log.Fatalf("merging user info: %v\n", err)
	}

	assetGroups, err := GetAssetGroup()
	if err != nil {
		log.Fatalf("fetching asset groups: %v\n", err)
	}

	asm, err := MergeAssetInfo(ea, cfg)
	if err != nil {
		log.Fatalf("merging asset info: %v\n", err)
	}

	tx, err := dao.DB.Beginx()
	if err != nil {
		log.Fatalf("beginning transaction: %v\n", err)
	}
	defer tx.Rollback()

	for _, u := range eusers {
		if _, err := tx.NamedExec(`update users set total_storage_size=:total_storage_size, used_storage_size=:used_storage_size, 
		api_keys=:api_keys, total_traffic=:total_traffic, peak_bandwidth=:peak_bandwidth, download_count=:download_count, enable_vip=:enable_vip where id=:id`, u); err != nil {
			log.Fatalf("updating user info: %v\n", err)
		}
	}

	for _, a := range asm {
		if _, err := tx.NamedExec(`update assets set group_id=:group_id, area_id=:area_id, share_status=:share_status, visit_count=:visit_count where hash=:hash`, a); err != nil {
			log.Fatalf("updating asset info: %v\n", err)
		}
	}

	for _, ag := range assetGroups {
		if _, err := tx.NamedExec(`insert into asset_group (id, user_id, name, parent, created_time) values (:id, :user_id, :name, :parent, :created_time)`, ag); err != nil {
			log.Fatalf("inserting asset group: %v\n", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commiting transaction: %v\n", err)
	}

}

/*
alter table users add column `total_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `used_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `total_traffic` BIGINT NOT NULL DEFAULT 0;
alter table users add column `peak_bandwidth` INT NOT NULL DEFAULT 0;
alter table users add column `download_count` INT NOT NULL DEFAULT 0;
alter table users add column `enable_vip` BOOLEAN DEFAULT false;
alter table users add column `api_keys` BLOB;
*/

type suser struct {
	user_id            string
	total_storage_size int64
	used_storage_size  int64
	api_keys           []byte
	total_traffic      int64
	peak_bandwidth     int64
	download_count     int64
	enable_vip         bool
}

func MergeUserInfo(em map[string]*model.User) (map[string]*model.User, error) {
	f, err := xlsx.OpenFile("./user_info.xlsx")
	if err != nil {
		return nil, err
	}

	var (
		sm = make(map[string]suser, 0)
	)

	susers := f.Sheets[0]
	if err := susers.ForEachRow(func(r *xlsx.Row) error {
		// 忽略第0行
		if r.GetCell(0).String() == "user_id" {
			return nil
		}
		user_id := r.GetCell(0).String()
		total_storage_size, err := r.GetCell(1).Int64()
		if err != nil {
			return err
		}
		used_storage_size, err := r.GetCell(2).Int64()
		if err != nil {
			return err
		}
		api_keys := r.GetCell(3).String()
		total_traffic, err := r.GetCell(4).Int64()
		if err != nil {
			return err
		}
		peak_bandwidth, err := r.GetCell(5).Int64()
		if err != nil {
			return err
		}
		download_count, err := r.GetCell(6).Int64()
		if err != nil {
			return err
		}
		enable_vip := r.GetCell(7).Bool()
		sm[user_id] = suser{
			user_id:            user_id,
			total_storage_size: total_storage_size,
			used_storage_size:  used_storage_size,
			api_keys:           []byte(api_keys),
			total_traffic:      total_traffic,
			peak_bandwidth:     peak_bandwidth,
			download_count:     download_count,
			enable_vip:         enable_vip,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ret = make(map[string]*model.User)
	for _, u := range sm {
		if _, ok := em[u.user_id]; !ok {
			continue
		}
		ret[u.user_id] = em[u.user_id]
		ret[u.user_id].TotalStorageSize = u.total_storage_size
		ret[u.user_id].UsedStorageSize = u.used_storage_size
		ret[u.user_id].TotalTraffic = u.total_traffic
		ret[u.user_id].PeakBandwidth = u.peak_bandwidth
		ret[u.user_id].DownloadCount = u.download_count
		ret[u.user_id].EnableVIP = u.enable_vip
		ret[u.user_id].ApiKeys = u.api_keys
	}

	return ret, nil
}

// ID          int64     `db:"id"`
//
//	UserID      string    `db:"user_id"`
//	Name        string    `db:"name"`
//	Parent      int64     `db:"parent"`
//	CreatedTime time.Time `db:"created_time"`
func GetAssetGroup() ([]*model.AssetGroup, error) {
	f, err := xlsx.OpenFile("./user_asset_group.xlsx")
	if err != nil {
		return nil, err
	}
	var ret []*model.AssetGroup
	ug := f.Sheets[0]
	if err := ug.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCell(0).String() == "id" {
			return nil
		}
		id, err := r.GetCell(0).Int64()
		if err != nil {
			return err
		}
		user_id := r.GetCell(1).String()
		name := r.GetCell(2).String()
		parent, err := r.GetCell(3).Int64()
		if err != nil {
			return err
		}
		created_time := r.GetCell(4).String()
		ct, err := time.Parse("2006-01-02 15:04:05", created_time)
		if err != nil {
			return err
		}
		ret = append(ret, &model.AssetGroup{
			ID:          id,
			UserID:      user_id,
			Name:        name,
			Parent:      parent,
			CreatedTime: ct,
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

// alter table assets add column `group_id` int NOT NULL DEFAULT 0;
// alter table assets add column `area_id` varchar(255) NOT NULL DEFAULT 0;
// alter table assets add column `share_status` TINYINT NOT NULL DEFAULT 0;
func MergeAssetInfo(ea map[string]*model.Asset, cfg config.Config) (map[string]*model.Asset, error) {
	f, err := xlsx.OpenFile("./user_asset.xlsx")
	if err != nil {
		return nil, err
	}

	as := f.Sheets[0]
	var ret = make(map[string]*model.Asset)

	if err := as.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCell(0).String() == "hash" {
			return nil
		}
		hash := r.GetCell(0).String()
		if _, ok := ea[hash]; !ok {
			return nil
		}
		group_id, err := r.GetCell(9).Int64()
		if err != nil {
			return err
		}
		area_id := cfg.SpecifyCandidate.AreaId
		share_status, err := r.GetCell(4).Int64()
		if err != nil {
			return err
		}
		ret[hash] = ea[hash]
		ret[hash].GroupID = group_id
		ret[hash].AreaID = area_id
		ret[hash].ShareStatus = share_status
		return nil
	}); err != nil {
		return nil, err
	}

	f, err = xlsx.OpenFile("./asset_visit_count.xlsx")
	if err != nil {
		return nil, err
	}

	av := f.Sheets[0]
	if err := av.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCell(0).String() == "hash" {
			return nil
		}
		hash := r.GetCell(0).String()
		if _, ok := ea[hash]; !ok {
			return nil
		}
		visit_count, err := r.GetCell(1).Int64()
		if err != nil {
			return err
		}
		if _, ok := ret[hash]; ok {
			ret[hash].VisitCount = visit_count
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}
