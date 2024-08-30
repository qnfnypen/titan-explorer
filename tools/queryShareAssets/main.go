package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/dao"
	"log"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/spf13/viper"
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

	etcdClient, err := statistics.NewEtcdClient(cfg.EtcdAddresses)
	if err != nil {
		log.Fatalf("New etcdClient Failed: %v", err)
	}

	schedulers, err := statistics.FetchSchedulersFromEtcd(etcdClient)
	if err != nil {
		log.Fatalf("fetch scheduler from etcd Failed: %v", err)
	}

	type downloadInfo struct {
		Url    string `json:"url"`
		AreaCN string `json:"area_cn"`
		AreaEn string `json:"area_en"`
	}

	type cliInfo struct {
		OS  string `json:"os"`
		Url string `json:"url"`
	}

	type response struct {
		Android []downloadInfo `json:"android"`
		MacOs   []downloadInfo `json:"macos"`
		Windows []downloadInfo `json:"windows"`
		Cli     []cliInfo      `json:"cli"`
	}

	type userAsset struct {
		UserId string
		Cid    string
	}

	out := &response{
		Android: make([]downloadInfo, 0),
		MacOs:   make([]downloadInfo, 0),
		Windows: make([]downloadInfo, 0),
		Cli: []cliInfo{
			{OS: "Linux", Url: "https://github.com/Titannet-dao/titan-node/releases/download/v0.1.20/titan-edge_v0.1.20_246b9dd_linux-amd64.tar.gz"},
			{OS: "MacOs (CLI)", Url: "https://github.com/Titannet-dao/titan-node/releases/download/v0.1.20/titan-edge_v0.1.20_246b9dd_mac_amd64.tar.gz"},
			{OS: "Windows (CLI)", Url: "https://github.com/Titannet-dao/titan-node/releases/download/v0.1.20/titan-edge_v0.1.20_246b9dd_widnows_amd64.tar.gz"},
			{OS: "Docker", Url: ""},
		},
	}

	cids := map[string]userAsset{
		"android": {UserId: "titan17ljevhtqu4vx6y7k743jyca0w8gyfu2466e8x3", Cid: "bafybeicznvslgyuhdwnrw5epabcp7nppbzkn6kjhcjumfb2ulhmay4pixy"},
		"mac":     {UserId: "0x7803c1e839a8101b37c90e42e440a837b192ae9e", Cid: "bafybeibz4nj72svea2goowncunmmukt3q67kfw4tvud52unkiutifpy5du"},
		"windows": {UserId: "0x7803c1e839a8101b37c90e42e440a837b192ae9e", Cid: "bafybeibry7lqb5soj52vl77fqp2wigbnwrklwaa5w77y2tvsthksldymsa"},
	}

	areasName := map[string]struct {
		CN string
		EN string
	}{
		"NorthAmerica-Canada":                  {EN: "Canada", CN: "加拿大"},
		"NorthAmerica-UnitedStates-California": {EN: "America", CN: "美国"},
		"Asia-Vietnam-Hanoi-Hanoi":             {EN: "Vietnam", CN: "越南"},
		"Europe-UnitedKingdom-England-London":  {EN: "England", CN: "英国"},
		"Asia-SouthKorea-Seoul-Seoul":          {EN: "SouthKorea", CN: "韩国"},
		"Asia-Japan-Tokyo-Tokyo":               {EN: "Japan", CN: "日本"},
		"Asia-Singapore":                       {EN: "Singapore", CN: "新加坡"},
		"Asia-China-Guangdong-Shenzhen":        {EN: "China", CN: "中国"},
		"Europe-Germany-Hesse-FrankfurtamMain": {EN: "Germany", CN: "德国"},
		"Asia-HongKong":                        {EN: "Common", CN: "公共"},
		"NorthAmerica-UnitedStates":            {EN: "America2", CN: "美国2"},
	}

	for _, schedulerClient := range schedulers {

		for area, ua := range cids {
			_, err := schedulerClient.Api.ShareAssets(context.Background(), ua.UserId, []string{ua.Cid}, time.Time{})
			if err != nil {
				continue
			}

			url := fmt.Sprintf(`https://api-test1.container1.titannet.io/api/v1/storage/open_asset?user_id=%s&asset_cid=%s&area_id=%s`, ua.UserId, ua.Cid, schedulerClient.AreaId)

			download := downloadInfo{
				Url:    url,
				AreaCN: areasName[schedulerClient.AreaId].CN,
				AreaEn: areasName[schedulerClient.AreaId].EN,
			}

			switch area {
			case "android":
				out.Android = append(out.Android, download)
			case "mac":
				out.MacOs = append(out.MacOs, download)
			case "windows":
				out.Windows = append(out.Windows, download)
			}
		}
	}

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(out); err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())

}
