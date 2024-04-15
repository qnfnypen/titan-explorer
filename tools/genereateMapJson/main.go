package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/spf13/viper"
	"log"
	"os"
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

	outEn, err := GetDeviceMapInfo(ctx, model.LanguageEN)
	if err != nil {
		log.Fatalln(err)
	}

	fileEn, _ := os.Create("map_en.json")
	defer fileEn.Close()

	bn, _ := json.Marshal(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"list": outEn,
		},
	})
	fileEn.Write(bn)

	//outCn, err := GetDeviceMapInfo(ctx, model.LanguageEN)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//fileEn, _ := os.Create("map_cn.json")
	//defer fileEn.Close()
	//
	//bn, _ := json.Marshal(outEn)
	//fileEn.Write(bn)

	log.Println("Success")
}

func GetDeviceMapInfo(ctx context.Context, lang model.Language) ([]*dao.MapInfo, error) {
	location := "location_en"
	if lang == model.LanguageCN {
		location = "location_cn"
	}

	var where string

	query := fmt.Sprintf(`select t.name, CONCAT(
    SUBSTRING_INDEX(t.external_ip, '.', 1), 
    '.xxx.xxx.', 
    SUBSTRING_INDEX(t.external_ip, '.', -1)
  ) AS ip, t.node_type, t.longitude, t.latitude from  (select IF(lc.city <> '', lc.city, lc.country) as name, external_ip , d.node_type, d.longitude, d.latitude from device_info d  
      left join %s lc on d.external_ip = lc.ip  where device_status_code = 1 and ip_country <> 'China' %s) t `, location, where)

	rows, err := dao.DB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []*dao.MapInfo

	for rows.Next() {
		var (
			name, nodeType, ip string
			lat, long          float64
		)

		if err := rows.Scan(&name, &ip, &nodeType, &long, &lat); err != nil {
			continue
		}

		out = append(out, &dao.MapInfo{
			Name:     name,
			NodeType: nodeType,
			Ip:       ip,
			Value:    []float64{lat, long},
		})
	}

	return out, nil
}
