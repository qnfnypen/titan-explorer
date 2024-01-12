package main

import (
	"context"
	"database/sql"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
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

	var locations []*model.Location
	if err := dao.DB.SelectContext(context.Background(), &locations, `SELECT * FROM location_cn `); err != nil {
		log.Fatal(err)
	}

	log.Printf("Preparing to parse %d locations\n", len(locations))

	var la model.Location
	ctx := context.Background()
	for i, location := range locations {
		log.Printf("Parse IP location %d/%d\n", i+1, len(locations))

		err := dao.DB.Get(&la, `SELECT * FROM location_en where ip = ?`, location.Ip)
		if err == nil {
			log.Printf("ip %s exist\n", location.Ip)
			continue
		}

		if err != sql.ErrNoRows {
			log.Println(err)
			continue
		}

		loc, err := iptool.IPDataCloudGetLocation(ctx, cfg.IpDataCloud.Url, location.Ip, cfg.IpDataCloud.Key, model.LanguageEN)
		if err != nil {
			log.Println(err)
			continue
		}

		err = dao.UpsertLocationInfo(ctx, loc, model.LanguageEN)
		if err != nil {
			log.Println(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Success")
}
