package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	"time"
)

const tableNameCacheEvent = "cache_event"

//func GetAreaID(ctx context.Context, userId string) string {
//
//	var areaID string
//	err := DB.GetContext(ctx, &areaID, fmt.Sprintf(
//		`SELECT area_id FROM %s where user_id = '%s' order by id desc limit 1`, tableNameApplication, userId,
//	))
//
//	if err == sql.ErrNoRows {
//		return GetLoginLocation(ctx, userId)
//	}
//
//	return areaID
//}

func CreateLink(ctx context.Context, link *model.Link) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (username, cid, short_link, long_link)
			VALUES (:username, :cid, :short_link, :long_link);`, tableNameLink,
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
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
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
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
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
		key := startTime.Format(formatter.TimeFormatYMDH)
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
	startTime, _ := time.Parse(formatter.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(formatter.TimeFormatDateOnly, end)
	oneDay := 24 * time.Hour
	deviceInDate := make(map[string]*CacheStatistics)
	var out []*CacheStatistics
	for _, data := range in {
		deviceInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatDateOnly)
		val, ok := deviceInDate[key]
		if !ok {
			val = &CacheStatistics{}
		}
		val.Date = startTime.Format(formatter.TimeFormatMD)
		out = append(out, val)
		startTime = startTime.Add(oneDay)
	}

	return out

}
