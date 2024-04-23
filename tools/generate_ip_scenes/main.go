package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx/v3"
	"io"
	"log"
	"net/http"
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
	file, err := xlsx.OpenFile("./singapore_ips.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	var count int

	sh := file.Sheets[0]
	err = sh.ForEachRow(func(r *xlsx.Row) error {
		ip := r.GetCell(0).String()

		fmt.Println(ip)

		count++

		scenes, err := IPDataCloudGetScenes(ctx, cfg.IpDataCloud.Url, ip, cfg.IpDataCloud.Key, model.LanguageCN)
		if err != nil {
			log.Println("query scenes", ip, err)
			return nil
		}

		scenes.IP = ip

		_, err = dao.DB.NamedExecContext(ctx, `insert into scenes(ip, asn, isp, usage_type) values (:ip, :asn, :isp, :usage_type)`, scenes)
		if err != nil {
			log.Println("insert err", err)
		}

		//fmt.Println(r.GetCell(0))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("handle %d done\n", count)

	log.Println("Success")
}

type result struct {
	Code int    `json:"code"`
	Data Data   `json:"data"`
	Msg  string `json:"msg"`
}

type Data struct {
	Scenes scenes `json:"scenes"`
}

type scenes struct {
	IP        string `json:"ip" db:"ip"`
	ASN       string `json:"asn" db:"asn"`
	ISP       string `json:"isp" db:"isp"`
	UsageType string `json:"usage_type" db:"usage_type"`
}

func IPDataCloudGetScenes(ctx context.Context, url, ip, key, lang string) (*scenes, error) {
	reqURL := fmt.Sprintf("%s?ip=%s&key=%s&language=%s", url, ip, key, lang)
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response: %d %v", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res result
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Msg)
	}

	return &res.Data.Scenes, nil
}
