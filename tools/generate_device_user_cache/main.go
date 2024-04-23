package main

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/spf13/viper"
	"log"
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

	loadDeviceUserToCache()
	//updateDeviceInfoDailyUser()

	//updateDeviceInfoUser()

	log.Println("Success")
}

//
//func updateDeviceInfoUser() {
//	log.Println("start to update device userid")
//
//	query := "select DISTINCT device_id from device_info where user_id = ''"
//	rows, err := dao.DB.QueryxContext(context.Background(), query)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	defer rows.Close()
//
//	var deviceId string
//	for rows.Next() {
//		err = rows.Scan(&deviceId)
//		if err != nil {
//			fmt.Println("scan: ", err)
//			continue
//		}
//
//		userId, err := dao.GetDeviceUserIdFromCache(context.Background(), deviceId)
//		if err != nil {
//			continue
//		}
//
//		if userId == "" {
//			continue
//		}
//
//		_, err = dao.DB.ExecContext(context.Background(), "update device_info set user_id = ? where device_id = ? and user_id = ''", userId, deviceId)
//		if err != nil {
//			fmt.Println("update userid: ", err)
//		}
//	}
//}

//func updateDeviceInfoDailyUser() {
//	log.Println("start to update device userid")
//
//	query := "select DISTINCT device_id from device_info_daily where user_id = ''"
//	rows, err := dao.DB.QueryxContext(context.Background(), query)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	defer rows.Close()
//
//	var deviceId string
//	for rows.Next() {
//		err = rows.Scan(&deviceId)
//		if err != nil {
//			fmt.Println("scan: ", err)
//			continue
//		}
//
//		userId, err := dao.GetDeviceUserIdFromCache(context.Background(), deviceId)
//		if err != nil {
//			continue
//		}
//
//		if userId == "" {
//			continue
//		}
//
//		_, err = dao.DB.ExecContext(context.Background(), "update device_info_daily set user_id = ? where device_id = ?", userId, deviceId)
//		if err != nil {
//			fmt.Println("update userid: ", err)
//		}
//	}
//}

func loadDeviceUserToCache() {
	//var users []*model.User
	log.Println("start to load device user ref")

	keyVal := make(map[string]map[string]string)
	rows, err := dao.DB.QueryxContext(context.Background(), "select device_id, area_id, user_id from device_info where user_id <> ''")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	var deviceId, userId, areaId string
	for rows.Next() {
		err = rows.Scan(&deviceId, &areaId, &userId)
		if err != nil {
			fmt.Println("scan: ", err)
			continue
		}

		if _, ok := keyVal[areaId]; !ok {
			keyVal[areaId] = make(map[string]string)
		}

		keyVal[areaId][deviceId] = userId
	}

	for key, val := range keyVal {
		err = dao.SetMultipleDeviceUserIdToCache(context.Background(), key, val)
		if err != nil {
			fmt.Println(err)
		}

		out, err := dao.RedisCache.HGetAll(context.Background(), fmt.Sprintf("TITAN::DEVICEUSERS::%s", key)).Result()
		if err != nil {
			fmt.Println("hget", err)
		}
		fmt.Println("==>>", len(out))
	}
}
