package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"time"
)

const tableNameCacheEvent = "cache_event"

func CreateCacheEvents(ctx context.Context, events []*model.CacheEvent) error {
	query := fmt.Sprintf(`INSERT INTO %s(device_id, carfile_cid,block_size,status, blocks, replicaInfos ,time) 
			VALUES(:device_id, :carfile_cid, :block_size, :status, :blocks, :replicaInfos, :time) ON DUPLICATE KEY UPDATE blocks = VALUES(blocks),
	block_size = VALUES(block_size),status = VALUES(status),replicaInfos = VALUES(replicaInfos),time = VALUES(time),updated_at = now()`, tableNameCacheEvent)

	_, err := DB.NamedExecContext(ctx, query, events)
	return err
}

func ResetCacheEvents(ctx context.Context, carFileCid string) error {
	_, err := DB.DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET status = 4, updated_at = now() WHERE carfile_cid = '%s'`, tableNameCacheEvent, carFileCid))
	return err
}

func GetReplicaInfo(ctx context.Context) model.FullNodeInfo {
	var ReplicaInfo model.FullNodeInfo
	query := fmt.Sprintf("SELECT count(*) as t_upstream_file_count,ROUND(count(*)/count(distinct(carfile_cid)),2) AS t_average_replica FROM %s where status = 3 ", tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&ReplicaInfo)
	if err != nil {
		log.Errorf("getDeviceInfo %v", err)
	}
	return ReplicaInfo
}
func GetAreaID(ctx context.Context, userId string) string {
	var areaID string
	_ = DB.GetContext(ctx, &areaID, fmt.Sprintf(
		`SELECT area_id FROM %s where user_id = '%s' order by id desc limit 1`, tableNameApplication, userId,
	))
	return areaID
}

func GetLastCacheEvent(ctx context.Context) (*model.CacheEvent, error) {
	var out model.CacheEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY time DESC LIMIT 1;`, tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func CountCacheEvent(ctx context.Context, nodeId string) error {
	var out model.DeviceInfo
	query := fmt.Sprintf(`SELECT device_id,count(DISTINCT(carfile_cid)) as block_count FROM %s where status = 3 and device_id = '%s' GROUP BY device_id;`, tableNameCacheEvent, nodeId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	err = UpdateCarFileCount(ctx, &out)
	if err != nil {
		return err
	}
	return nil
}

func QueryCacheHour(deviceID, startTime, endTime string) []*CacheStatistics {
	option := QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if option.StartTime == "" {
		option.StartTime = carbon.Now().StartOfHour().SubHours(25).String()
	}
	if option.EndTime == "" {
		option.EndTime = carbon.Now().String()
	} else {
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}
	where := `WHERE status = 3 `
	if deviceID != "" {
		where += fmt.Sprintf(" AND device_id = '%s'", deviceID)
	}
	list, err := GetCacheInfoHourList(context.Background(), where, option)
	if err != nil {
		log.Errorf("get incoming hour daily: %v", err)
		return nil
	}

	return list
}

func QueryCacheDaily(deviceID, startTime, endTime string) []*CacheStatistics {
	option := QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if startTime == "" {
		option.StartTime = carbon.Now().SubDays(14).StartOfDay().String()
	}
	if endTime == "" {
		option.EndTime = carbon.Now().EndOfDay().String()
	} else {
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}
	where := `WHERE status = 3 `
	if deviceID != "" {
		where += fmt.Sprintf(" AND device_id = '%s'", deviceID)
	}
	list, err := GetCacheInfoDaysList(context.Background(), where, option)
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
		return nil
	}

	return list
}

type CacheStatistics struct {
	Date       string  `json:"date" db:"date"`
	BlockCount float64 `json:"block_count" db:"block_count"`
}

func GetCacheInfoHourList(ctx context.Context, where string, option QueryOption) ([]*CacheStatistics, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d %%H') as date, COUNT(1) as block_count
	from %s %s and time>='%s' and time<='%s' group by date order by date`, tableNameCacheEvent, where, option.StartTime, option.EndTime)
	var out []*CacheStatistics
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}
	if len(out) > 0 {
		return handleCacheHourList(out[1:]), err
	}
	return out, err
}

func GetCacheInfoDaysList(ctx context.Context, where string, option QueryOption) ([]*CacheStatistics, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d') as date, COUNT(1) as block_count
	from %s %s and time>='%s' and time<='%s' group by date order by date`, tableNameCacheEvent, where, option.StartTime, option.EndTime)
	var out []*CacheStatistics
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}
	if len(out) > 0 {
		return handleCacheDaysList(option.StartTime[0:10], option.EndTime[0:10], out), err
	}
	return out, err
}

func handleCacheHourList(in []*CacheStatistics) []*CacheStatistics {
	now := time.Now()
	oneHour := time.Hour
	startTime := now.Add(-23 * oneHour)
	endTime := now
	deviceInDate := make(map[string]*CacheStatistics)
	var out []*CacheStatistics
	for _, data := range in {
		deviceInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(utils.TimeFormatYMDH)
		val, ok := deviceInDate[key]
		if !ok {
			val = &CacheStatistics{}
		}
		val.Date = fmt.Sprintf("%d:00", startTime.Hour())
		out = append(out, val)
		startTime = startTime.Add(oneHour)
	}

	return out
}

func handleCacheDaysList(start, end string, in []*CacheStatistics) []*CacheStatistics {
	startTime, _ := time.Parse(utils.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(utils.TimeFormatDateOnly, end)
	oneDay := 24 * time.Hour
	deviceInDate := make(map[string]*CacheStatistics)
	var out []*CacheStatistics
	for _, data := range in {
		deviceInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(utils.TimeFormatDateOnly)
		val, ok := deviceInDate[key]
		if !ok {
			val = &CacheStatistics{}
		}
		val.Date = startTime.Format(utils.TimeFormatMD)
		out = append(out, val)
		startTime = startTime.Add(oneDay)
	}

	return out

}

func GetCacheEventsByPage(ctx context.Context, cond *model.CacheEvent, option QueryOption) ([]*model.CacheEvent, int64, error) {
	var args []interface{}
	where := `WHERE 1=1 AND status=3`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	} else {
		where += " ORDER BY time DESC"
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
	var out []*model.CacheEvent

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameCacheEvent, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameCacheEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
