package main

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/geo"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	"github.com/spf13/viper"
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

	var ips []string
	if err := dao.DB.SelectContext(context.Background(), &ips, `select distinct external_ip from device_info where external_ip <> ''`); err != nil {
		log.Fatal(err)
	}

	log.Printf("Preparing to parse %d locations\n", len(ips))

	updateStatement := `update device_info set ip_location = ?, ip_country = ?, ip_province = ?, ip_city = ?, longitude = ?, latitude = ?, updated_at = now() where external_ip = ?`
	
	ctx := context.Background()
	for _, ip := range ips {
		log.Println("query ip: ", ip)

		locEn, err := iptool.IPDataCloudGetLocation(ctx, cfg.IpDataCloud.Url, ip, cfg.IpDataCloud.Key, model.LanguageEN)
		if err != nil {
			log.Println("get location cn: ", err)
			continue
		}

		fmt.Println("ip: =>", dao.ContactIPLocation(*locEn, model.LanguageEN))

		err = dao.UpsertLocationInfo(ctx, locEn, model.LanguageEN)
		if err != nil {
			log.Println("update location en: ", err)
		}

		_, err = dao.DB.ExecContext(ctx, updateStatement, dao.ContactIPLocation(*locEn, model.LanguageEN), locEn.Country, locEn.Province, locEn.City, locEn.Longitude, locEn.Latitude, ip)
		if err != nil {
			log.Println("update deviceInfo err ip: ", ip, " err", err)
			continue
		}

		err = geo.CacheIPLocation(ctx, locEn, model.LanguageEN)
		if err != nil {
			log.Println("cache en", err)
		}

		locCn, err := iptool.IPDataCloudGetLocation(ctx, cfg.IpDataCloud.Url, ip, cfg.IpDataCloud.Key, model.LanguageCN)
		if err != nil {
			log.Println("get location cn: ", err)
			continue
		}

		err = dao.UpsertLocationInfo(ctx, locCn, model.LanguageCN)
		if err != nil {
			log.Println("update location cn: ", err)
		}

		err = geo.CacheIPLocation(ctx, locCn, model.LanguageCN)
		if err != nil {
			log.Println("cache cn:", err)
		}

		log.Println("update ", ip, "success")
		time.Sleep(10 * time.Millisecond)
	}

	log.Println("Success")
}
