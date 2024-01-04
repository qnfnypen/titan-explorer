package main

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/random"
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

	ctx := context.Background()

	rows, err := dao.DB.QueryxContext(ctx, "select * from users")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	updateStatement := "update users set referral_code = ? where username = ? and referral_code = ''"

	for rows.Next() {

		var u model.User
		if err := rows.StructScan(&u); err != nil {
			log.Println("sql rows scan err: ", err)
			continue
		}

		referralCode := random.GenerateRandomString(6)

		_, err = dao.DB.ExecContext(ctx, updateStatement, referralCode, u.Username)
		if err != nil {
			log.Println("sql rows update refferal code err: ", err)
		}
	}

	log.Println("Success")
}
