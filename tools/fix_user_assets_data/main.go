package main

import (
	"context"
	"encoding/csv"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"strconv"
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

	file, err := os.Open("./user_asset.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)

	var asssets []*model.Asset

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		totalSize, _ := strconv.ParseInt(record[4], 10, 64)

		asssets = append(asssets, &model.Asset{
			Hash:       record[0],
			Cid:        record[1],
			CreatedAt:  carbon.Parse(record[2]).Carbon2Time(),
			Name:       record[3],
			TotalSize:  totalSize,
			Type:       record[5],
			Expiration: carbon.Parse(record[7]).Carbon2Time(),
		})
	}

	err = dao.AddAssets(context.Background(), asssets)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Success")
}
