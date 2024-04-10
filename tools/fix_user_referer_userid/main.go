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

	updateUserRefererUserId()

	log.Println("Success")
}

func updateUserRefererUserId() {
	log.Println("start to update referrer userid")

	query := "select ifnull(u2.username,'') as user_id, u1.username as referrer  from users u1 inner join users u2 on u1.referral_code = u2.referrer"
	rows, err := dao.DB.QueryxContext(context.Background(), query)
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	var userId, referrerUserId string
	for rows.Next() {
		err = rows.Scan(&userId, &referrerUserId)
		if err != nil {
			fmt.Println("scan: ", err)
			continue
		}

		if userId == "" || referrerUserId == "" {
			continue
		}

		_, err = dao.DB.ExecContext(context.Background(), "update users set referrer_user_id = ? where username = ?", referrerUserId, userId)
		if err != nil {
			fmt.Println("update userid: ", err)
		}
	}
}
