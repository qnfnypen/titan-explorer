package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	tapi "github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/node/cidutil"
	"github.com/gnasnik/titan-explorer/api"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Detrail struct {
	Hash   string `db:"hash" gorm:"column:ha"`
	UserID string `db:"user_id" gorm:"column:uid"`
	AreaID string `db:"area_id" gorm:"column:aid"`
	IsSync bool   `db:"is_sync" gorm:"column:is_sync"`
	CID    string `gorm:"column:cid"`
}

type UserAssetArea struct {
	Hash   string `db:"hash" gorm:"column:hash"`
	UserID string `db:"user_id" gorm:"column:user_id"`
	AreaID string `db:"area_id" gorm:"column:area_id"`
	IsSync bool   `db:"is_sync" gorm:"column:is_sync"`
}

var (
	wg       = new(sync.WaitGroup)
	aidMaps  = make(map[string]tapi.Scheduler)
	l1States = []string{"EdgesSelect", "EdgesPulling", "Servicing", "EdgesFailed"}
	uid      = "titan16k6k5lquc7c8pjcc8u2y3kw72x9qnvnqsnmwvg"
	// dsn := fmt.Sprintf("%s:%s@tcp(localhost:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, localPort, dbName)
	dsn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:8080)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "nH9NWeucerLE56Jv", "titan_explorer")
	db  *gorm.DB
	ctx = context.Background()
)

func Init() {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func (UserAssetArea) TableName() string {
	return "user_asset_area"
}

func main() {
	var err error
	Init()

	// export ETCD_USERNAME=web
	// export ETCD_PASSWORD=web_123
	os.Setenv("ETCD_USERNAME", "web")
	os.Setenv("ETCD_PASSWORD", "web_123")

	uas := getUserAssetAreas(uid)
	wg.Add(len(uas))
	stockChan := make(chan struct{}, 50)

	for i, v := range uas {
		go func(i int, v Detrail) {
			stockChan <- struct{}{}
			defer func() {
				log.Println(i)
				wg.Done()
				<-stockChan
			}()
			scli := getAreaSlic(v.AreaID)
			if v.CID == "" {
				if v.CID, err = cidutil.HashToCID(v.Hash); err != nil {
					log.Printf("hash to cid error:%v\n", err)
					return
				}
			}
			rs, err := scli.GetAssetRecord(ctx, v.CID)
			if err != nil {
				log.Println(err)
				updateUserAssetArea(uid, v.Hash, v.AreaID, false)
				return
			}
			if checkSyncState(rs.State) {
				updateUserAssetArea(uid, v.Hash, v.AreaID, true)
			} else {
				updateUserAssetArea(uid, v.Hash, v.AreaID, false)
			}
		}(i, v)
	}
	wg.Wait()
}

func getUserAssetAreas(uid string) []Detrail {
	var uas []Detrail

	err := db.Model(&UserAssetArea{}).Select("user_asset_area.hash AS ha,user_asset_area.user_id AS uid,user_asset_area.area_id AS aid,is_sync,cid").
		Joins("LEFT JOIN user_asset ON user_asset.hash = user_asset_area.hash").Where("user_asset_area.user_id = ?", uid).Scan(&uas).Error
	if err != nil {
		panic(err)
	}

	return uas
}

func updateUserAssetArea(uid, hash, aid string, isSync bool) {
	err := db.Model(&UserAssetArea{}).Where("user_id = ? AND hash = ? AND area_id = ?", uid, hash, aid).UpdateColumn("is_sync", isSync).Error
	if err != nil {
		log.Println(err)
	}
}

func checkSyncState(state string) bool {
	for _, v := range l1States {
		if strings.EqualFold(v, state) {
			return true
		}
	}

	return false
}

func getAreaSlic(areaID string) tapi.Scheduler {
	if v, ok := aidMaps[areaID]; ok {
		return v
	}

	scli, err := api.GetSchedulerClient(ctx, areaID)
	if err != nil {
		panic(err)
	}

	aidMaps[areaID] = scli
	return scli
}
