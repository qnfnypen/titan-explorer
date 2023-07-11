package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"time"
)

const tableNameCacheEvent = "cache_event"

func GetAreaID(ctx context.Context, userId string) string {
	var areaID string
	_ = DB.GetContext(ctx, &areaID, fmt.Sprintf(
		`SELECT area_id FROM %s where user_id = '%s' order by id desc limit 1`, tableNameApplication, userId,
	))
	return areaID
}

func CreateLink(ctx context.Context, link *model.Link) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (cid, short_link, long_link)
			VALUES (:cid, :short_link, :long_link);`, tableNameLink,
	), link)
	return err
}

func GetShortLink(ctx context.Context, link string) string {
	var areaID string
	_ = DB.GetContext(ctx, &areaID, fmt.Sprintf(
		`SELECT short_link FROM %s where long_link = '%s' order by id desc limit 1`, tableNameLink, link,
	))
	return areaID
}
func GetLongLink(ctx context.Context, cid string) string {
	var areaID string
	_ = DB.GetContext(ctx, &areaID, fmt.Sprintf(
		`SELECT long_link FROM %s where cid = '%s' order by id desc limit 1`, tableNameLink, cid,
	))
	return areaID
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
