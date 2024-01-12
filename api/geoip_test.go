package api

import (
	"context"
	"testing"

	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/spf13/viper"
)

func TestGeoIP(t *testing.T) {
	viper.AddConfigPath("../")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("reading config file: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unmarshaling config file: %v\n", err)
	}

	config.Cfg = cfg
	// if cfg.Mode == "debug" {
	// 	logging.SetDebugLogging()
	// }

	if err := dao.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	SchedulerConfigs = make(map[string][]*types.SchedulerCfg)
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJ3ZWIiXSwiSUQiOiIzOTkzN2E2Zi1lYjhmLTQxZjMtYjNlYS0xNWZlYjM3N2FjNjQiLCJOb2RlSUQiOiIiLCJFeHRlbmQiOiIifQ.6Sonm9R9pY6anX5iuX8QIfb46UMBN0Eltnx3_CtpL2M"
	schedulerCfg := &types.SchedulerCfg{SchedulerURL: "https://39.108.214.29:3456/rpc/v0", AreaID: "Asia-China-Guangdong-Shenzhen", AccessToken: accessToken}
	SchedulerConfigs["Asia-China-Guangdong-Shenzhen"] = []*types.SchedulerCfg{schedulerCfg}
	schedulerClient, err := getSchedulerClient(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	nodeIPInfos, err := schedulerClient.GetCandidateIPs(context.Background())
	if err != nil {
		t.Logf("get candidate ips error %s", err.Error())
	}

	if len(nodeIPInfos) == 0 {
		t.Logf("len(nodeIPInfos) == 0")
		return
	}

	userIP := "104.28.217.212"
	var nearestNode string
	if len(nodeIPInfos) > 0 {
		nodeMap := make(map[string]string)
		ips := make([]string, 0, len(nodeIPInfos))
		for _, nodeIPInfo := range nodeIPInfos {
			t.Logf("node %s %s", nodeIPInfo.NodeID, nodeIPInfo.IP)
			ips = append(ips, nodeIPInfo.IP)
			nodeMap[nodeIPInfo.IP] = nodeIPInfo.NodeID
		}

		if ip, err := GetUserNearestIP(context.Background(), userIP, ips, NewIPCoordinate()); err == nil {
			nearestNode = nodeMap[ip]
		} else {
			t.Logf("err:%s", err.Error())
		}
	}

	// ipList := []string{"183.60.189.250", "8.209.196.80", "8.213.145.168", "8.213.197.249", "47.91.43.78", "8.208.33.62", "47.254.153.80", "8.217.116.27", "47.251.59.211", "8.219.152.148"}
	// ip := GetUserNearestIP(context.Background(), userIP, ipList, NewIPCoordinate())
	t.Logf("nearestNode:%s", nearestNode)
}
