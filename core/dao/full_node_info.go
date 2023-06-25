package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/go-redis/redis/v9"
	"time"
)

var tableNameFullNodeInfo = "full_node_info"

const (
	FullNodeInfoKeyExpiration = 0
	FullNodeInfoKey           = "titan::full_node_info"
)

func UpsertFullNodeInfo(ctx context.Context, fullNodeInfo *model.FullNodeInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (t_node_online_ratio, t_average_replica, t_upstream_file_count, validator_count, candidate_count, edge_count, total_storage,storage_used, total_upstream_bandwidth, 
                total_downstream_bandwidth, total_carfile, total_carfile_size, retrieval_count, total_node_count,online_node_count, next_election_time, 
                time, created_at) 
		VALUES (:t_node_online_ratio, :t_average_replica, :t_upstream_file_count, :validator_count, :candidate_count, :edge_count, :total_storage, :storage_used, :total_upstream_bandwidth, :total_downstream_bandwidth,
		 :total_carfile, :total_carfile_size, :retrieval_count, :total_node_count, :online_node_count, :next_election_time, :time, :created_at) 
		 ON DUPLICATE KEY UPDATE t_node_online_ratio = VALUES(t_node_online_ratio), t_average_replica = VALUES(t_average_replica), t_upstream_file_count = VALUES(t_upstream_file_count), validator_count = VALUES(validator_count), candidate_count = VALUES(candidate_count),
		edge_count = VALUES(edge_count), total_storage = VALUES(total_storage), storage_used = VALUES(storage_used), total_upstream_bandwidth = VALUES(total_upstream_bandwidth),
		total_downstream_bandwidth = VALUES(total_downstream_bandwidth), total_carfile = VALUES(total_carfile), 
		total_carfile_size = VALUES(total_carfile_size), retrieval_count = VALUES(retrieval_count), total_node_count = VALUES(total_node_count), online_node_count = VALUES(online_node_count)`, tableNameFullNodeInfo),
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

type FullNodeDaysInfo struct {
	Date                     string  `json:"date" db:"date"`
	TotalNodeCount           int64   `db:"total_node_count" json:"total_node_count"`
	OnlineNodeCount          int64   `db:"online_node_count" json:"online_node_count"`
	FNodeCount               int64   `db:"f_node_count" json:"f_node_count"`
	VCCount                  int64   `db:"VC_count" json:"VC_count"`
	EdgeCount                int64   `db:"edge_count" json:"edge_count"`
	TUpstreamFileCount       int64   `db:"t_upstream_file_count" json:"t_upstream_file_count"`
	TotalStorage             float64 `db:"total_storage" json:"total_storage"`
	StorageUsed              float64 `db:"storage_used" json:"storage_used"`
	TotalUpstreamBandwidth   float64 `db:"total_upstream_bandwidth" json:"total_upstream_bandwidth"`
	TotalDownstreamBandwidth float64 `db:"total_downstream_bandwidth" json:"total_downstream_bandwidth"`
}

func QueryNodesDailyInfo(startTime, endTime string) []*FullNodeDaysInfo {
	list, err := GetNodesDaysList(context.Background(), OptionHandle(startTime, endTime))
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
		return nil
	}

	return list
}

func GetNodesDaysList(ctx context.Context, option QueryOption) ([]*FullNodeDaysInfo, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d') as date, 
	total_node_count,online_node_count,total_upstream_bandwidth,total_downstream_bandwidth,total_storage,storage_used,t_upstream_file_count,
	edge_count,(validator_count+candidate_count) as VC_count,f_node_count from %s where time>='%s' and time<='%s' group by date order by date ASC`, tableNameFullNodeInfo, option.StartTime, option.EndTime)
	var out []*FullNodeDaysInfo
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}
	if len(out) > 0 {
		return handleNodesDaysList(option.StartTime[0:10], option.EndTime[0:10], out), err
	}
	return out, err
}

func handleNodesDaysList(start, end string, in []*FullNodeDaysInfo) []*FullNodeDaysInfo {
	startTime, _ := time.Parse(utils.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(utils.TimeFormatDateOnly, end)
	oneDay := 24 * time.Hour
	deviceInDate := make(map[string]*FullNodeDaysInfo)
	var out []*FullNodeDaysInfo
	for _, data := range in {
		deviceInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(utils.TimeFormatDateOnly)
		val, ok := deviceInDate[key]
		if !ok {
			val = &FullNodeDaysInfo{}
		}
		val.Date = startTime.Format(utils.TimeFormatMD)
		out = append(out, val)
		startTime = startTime.Add(oneDay)
	}

	return out

}
