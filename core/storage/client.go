package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/node/scheduler"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/go-redis/redis/v9"
)

var (
	DefaultAreaId            = "Asia-China-Guangdong-Shenzhen"
	SchedulerConfigKeyPrefix = "TITAN::SCHEDULERCFG"
)

// Client 客户端
type Client struct {
	sc *scheduler.Scheduler
}

// NewClient 新建客户端
// func NewClient(areaID string) (*Client, error) {
// 	schedulerClient, err := getSchedulerClient(context.Background(), areaID)
// 	if err != nil {
// 		return nil, fmt.Errorf("new storage client error:%w", err)
// 	}

// 	sc, _ := schedulerClient.(*scheduler.Scheduler)

// 	return &Client{sc: sc}, nil
// }

// getSchedulerClient 获取调度器的 rpc 客户端实例, titan 节点是有区域区分的,不同的节点会连接不同区域的调度器,当需要查询该节点的数据时,需要连接对应的调度器
// areaId 区域Id在同步的节点的时候会写入到 device_info表,可以查询节点的信息,获得对应的区域ID,如果没有传区域ID,那么会遍历所有的调度器,可能会有性能问题.
func getSchedulerClient(ctx context.Context, areaId string) (api.Scheduler, error) {
	schedulers, err := statistics.GetSchedulerConfigs(ctx, fmt.Sprintf("%s::%s", SchedulerConfigKeyPrefix, areaId))
	if err == redis.Nil && areaId != DefaultAreaId {
		return getSchedulerClient(ctx, DefaultAreaId)
	}

	if err != nil || len(schedulers) == 0 {
		return nil, errors.New("no scheduler found")
	}

	schedulerApiUrl := schedulers[0].SchedulerURL
	schedulerApiToken := schedulers[0].AccessToken
	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+schedulerApiToken)
	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
	if err != nil {
		return nil, fmt.Errorf("create scheduler rpc client: %w", err)
	}

	return schedulerClient, nil
}
