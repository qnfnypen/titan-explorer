package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/go-redis/redis/v9"
)

var tableNameFullNodeInfo = "full_node_info"

const (
	FullNodeInfoKeyExpiration = 0
	FullNodeInfoKey           = "titan::full_node_info"
)

func UpsertFullNodeInfo(ctx context.Context, fullNodeInfo *model.FullNodeInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (validator_count, candidate_count, edge_count, total_storage, total_upstream_bandwidth, 
                total_downstream_bandwidth, total_carfile, total_carfile_size, retrieval_count, total_node_count, next_election_time, 
                time, created_at) 
		VALUES (:validator_count, :candidate_count, :edge_count, :total_storage, :total_upstream_bandwidth, :total_downstream_bandwidth,
		 :next_election_time, :total_carfile, :total_carfile_size, :retrieval_count, :total_node_count, :time, :created_at)`, tableNameFullNodeInfo),
		fullNodeInfo)
	return err
}

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
