package dao

import (
	"context"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/go-redis/redis/v9"
)

var tableNameFullNodeInfoHours = "full_node_info"

const (
	FullNodeInfoKeyExpiration = 0
	FullNodeInfoKey           = "titan::full_node_info"
)

func CacheFullNodeInfo(ctx context.Context, fullNodeInfo *model.FullNodeInfo) error {
	bytes, err := json.Marshal(fullNodeInfo)
	if err != nil {
		return err
	}
	_, err = Cache.Set(ctx, FullNodeInfoKey, bytes, FullNodeInfoKeyExpiration).Result()
	return err
}

func GetCacheFullNodeInfo(ctx context.Context) (*model.FullNodeInfo, error) {
	out := &model.FullNodeInfo{}
	bytes, err := Cache.Get(ctx, FullNodeInfoKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
