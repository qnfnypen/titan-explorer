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
		`INSERT INTO %s (t_node_online_ratio, t_average_replica, t_upstream_file_count, validator_count, candidate_count, edge_count, total_storage,storage_used, total_upstream_bandwidth, 
                total_downstream_bandwidth, total_carfile, total_carfile_size, retrieval_count, total_node_count, next_election_time, 
                time, created_at) 
		VALUES (:t_node_online_ratio, :t_average_replica, :t_upstream_file_count, :validator_count, :candidate_count, :edge_count, :total_storage, :storage_used, :total_upstream_bandwidth, :total_downstream_bandwidth,
		 :total_carfile, :total_carfile_size, :retrieval_count, :total_node_count, :next_election_time, :time, :created_at) 
		 ON DUPLICATE KEY UPDATE t_node_online_ratio = VALUES(t_node_online_ratio), t_average_replica = VALUES(t_average_replica), t_upstream_file_count = VALUES(t_upstream_file_count), validator_count = VALUES(validator_count), candidate_count = VALUES(candidate_count),
		edge_count = VALUES(edge_count), total_storage = VALUES(total_storage), storage_used = VALUES(storage_used), total_upstream_bandwidth = VALUES(total_upstream_bandwidth),
		total_downstream_bandwidth = VALUES(total_downstream_bandwidth), total_carfile = VALUES(total_carfile), 
		total_carfile_size = VALUES(total_carfile_size), retrieval_count = VALUES(retrieval_count), total_node_count = VALUES(total_node_count)`, tableNameFullNodeInfo),
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

func GetFullNodeInfoList(ctx context.Context, cond *model.FullNodeInfo, option QueryOption) ([]*model.FullNodeInfo, int64, error) {
	var args []interface{}
	where := `WHERE 1 = 1`
	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	} else {
		where += fmt.Sprintf(" ORDER BY time DESC")
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.FullNodeInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameFullNodeInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameFullNodeInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
