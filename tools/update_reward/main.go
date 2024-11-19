package main

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx/v3"
	"log"
	"time"
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

	userInDevice, err := dao.GetAllDeviceUserIdFromCache(ctx)
	if err != nil {
		log.Fatalf("get all device user id from cache: %v", err)
	}

	etcdClient, err := statistics.NewEtcdClient([]string{cfg.EtcdAddress})
	if err != nil {
		log.Fatalf("New etcdClient Failed: %v", err)
	}

	schedulers, err := statistics.FetchSchedulersFromEtcd(etcdClient)
	if err != nil {
		log.Fatalf("fetch scheduler from etcd Failed: %v", err)
	}

	sm := make(map[string]*statistics.Scheduler)

	for _, s := range schedulers {
		sm[s.AreaId] = s
	}

	file, err := xlsx.OpenFile("./sv.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	var count int

	var deviceInfos []*model.DeviceInfo

	sh := file.Sheets[1]
	err = sh.ForEachRow(func(r *xlsx.Row) error {
		nodeId := r.GetCell(0).String()
		areaId := r.GetCell(5).String()

		scheduler, ok := sm[areaId]
		if !ok {
			log.Println("not found", nodeId, areaId)
			return nil
		}

		resp, err := scheduler.Api.GetNodeInfo(ctx, nodeId)
		if err != nil {
			log.Printf("api GetNodeList from %s: %v\n", scheduler.AreaId, err)
			return nil
		}

		if resp.LastSeen.Add(14 * time.Hour).After(time.Now()) {
			return nil
		}

		count++

		fmt.Println(nodeId)

		deviceInfo := statistics.ToDeviceInfo(resp, areaId)

		userId, ok := userInDevice[deviceInfo.DeviceID]
		if !ok || userId == "" {
			userId = statistics.GetDeviceUserId(ctx, deviceInfo.DeviceID)
		}

		deviceInfos = append(deviceInfos, deviceInfo)

		if len(deviceInfos) > 1000 {
			log.Println("update device infos")
			if err := update(ctx, deviceInfos); err != nil {
				log.Println("update: ", err)
				return err
			}
			deviceInfos = make([]*model.DeviceInfo, 0)
		}

		//fmt.Println(r.GetCell(0))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if err := update(ctx, deviceInfos); err != nil {
		log.Fatal("update: ", err)
	}

	fmt.Printf("handle %d done\n", count)

	log.Println("Success")
}

func update(ctx context.Context, deviceInfos []*model.DeviceInfo) error {
	if len(deviceInfos) == 0 {
		return nil
	}

	err := dao.BulkUpsertDeviceInfo(ctx, deviceInfos)
	if err != nil {
		log.Printf("bulk upsert device info: %v\n", err)
		return err
	}

	start := time.Now()

	var deviceInfoHour []*model.DeviceInfoHour
	for _, d := range deviceInfos {
		deviceInfoHour = append(deviceInfoHour, statistics.ToDeviceInfoHour(d, start))
	}

	if err = statistics.AddDeviceInfoHours(ctx, start, deviceInfoHour); err != nil {
		log.Printf("add device info hours: %v", err)
	}

	if err := statistics.SumDailyReward(ctx, start, deviceInfos); err != nil {
		log.Printf("add device info daily reward: %v", err)
	}

	return nil
}
