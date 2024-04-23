package main

import (
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
	//
	//ctx := context.Background()
	//
	//userReward, err := dao.GetSumUserDeviceReward(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//userReferrerReward, err := dao.SumUserReferralReward2(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//dao.GetUserReferralReward()

	log.Println("Success")
}
