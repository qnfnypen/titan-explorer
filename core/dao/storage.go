package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/jmoiron/sqlx"
	"time"
)

const tableNameStorageHour = "storage_hour"

func BulkUpsertStorageHours(ctx context.Context, userInfos []*model.UserInfo) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, user_id,
				total_storage_size,used_storage_size,total_bandwidth,peak_bandwidth,download_count, time)
			VALUES (:created_at, :updated_at, :user_id, :total_storage_size, :used_storage_size, :total_bandwidth, :peak_bandwidth, :download_count, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), total_storage_size = VALUES(total_storage_size),used_storage_size = VALUES(used_storage_size),
			 total_bandwidth = VALUES(total_bandwidth),peak_bandwidth = VALUES(peak_bandwidth),download_count = VALUES(download_count)`, tableNameStorageHour)
	_, err := DB.NamedExecContext(ctx, upsertStatement, userInfos)
	return err
}

type UserInfoRes struct {
	Date           string `json:"date" db:"date"`
	TotalBandwidth int64  `db:"total_bandwidth"`
	PeakBandwidth  int64  `db:"peak_bandwidth"`
	DownloadCount  int64  `db:"download_count"`
}

func GetStorageInfoHourList(ctx context.Context, userId string, option QueryOption) ([]*UserInfoRes, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d %%H') as date,
	max(total_bandwidth) - min(total_bandwidth) as total_bandwidth,
	ifnull(max(peak_bandwidth),0) as peak_bandwidth, 
	max(download_count) - min(download_count) as download_count 
	from %s where user_id='%s' and time>='%s' and time<='%s' group by date order by date`, tableNameStorageHour, userId, option.StartTime, option.EndTime)
	var out []*UserInfoRes
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return out, nil
	}

	return handleStorageHourList(out), err
}

func handleStorageHourList(in []*UserInfoRes) []*UserInfoRes {
	now := time.Now()
	oneHour := time.Hour
	startTime := now.Add(-23 * oneHour)
	endTime := now
	userInDate := make(map[string]*UserInfoRes)
	var out []*UserInfoRes
	for _, data := range in {
		userInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatYMDH)
		val, ok := userInDate[key]
		if !ok {
			val = &UserInfoRes{}
		}
		val.Date = fmt.Sprintf("%d:00", startTime.Hour())
		out = append(out, val)
		startTime = startTime.Add(oneHour)
	}

	return out
}

func GetStorageInfoDaysList(ctx context.Context, userId string, option QueryOption) ([]*UserInfoRes, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d') as date,
	max(total_bandwidth) - min(total_bandwidth) as total_bandwidth,
	ifnull(max(peak_bandwidth),0) as peak_bandwidth, 
	max(download_count) - min(download_count) as download_count 
	from %s where user_id='%s' and time>='%s' and time<='%s' group by date order by date`, tableNameStorageHour, userId, option.StartTime, option.EndTime)
	var out []*UserInfoRes
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return out, nil
	}

	return handleStorageDailyList(option.StartTime, option.EndTime, out), err
}

func handleStorageDailyList(start, end string, in []*UserInfoRes) []*UserInfoRes {
	startTime, _ := time.Parse(formatter.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(formatter.TimeFormatDateOnly, end)
	oneDay := 24 * time.Hour
	deviceInDate := make(map[string]*UserInfoRes)
	var out []*UserInfoRes
	for _, data := range in {
		deviceInDate[data.Date] = data
	}

	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatDateOnly)
		val, ok := deviceInDate[key]
		if !ok {
			val = &UserInfoRes{}
		}
		val.Date = startTime.Format(formatter.TimeFormatMD)
		out = append(out, val)
		startTime = startTime.Add(oneDay)
	}

	return out

}

func GetAreaCount(ctx context.Context, deviceIds []string) (int64, error) {
	var CountAreas int64
	query, args, err := sqlx.In(fmt.Sprintf(
		`SELECT count(distinct(ip_city)) FROM %s WHERE device_id IN (?)`, tableNameDeviceInfo), deviceIds)
	if err != nil {
		return 0, err
	}
	query = DB.Rebind(query)
	err = DB.GetContext(ctx, &CountAreas, query, args...)
	if err != nil {
		return 0, err
	}
	return CountAreas, nil
}

func GetPeakBandwidth(ctx context.Context, userId string) (int64, error) {
	var peakBandwidth int64
	query, args, err := sqlx.In(fmt.Sprintf(
		`select ifnull(max(peak_bandwidth) ,0) as peak_bandwidth from %s where user_id='%s'`, tableNameStorageHour, userId))
	if err != nil {
		return 0, err
	}
	query = DB.Rebind(query)
	err = DB.GetContext(ctx, &peakBandwidth, query, args...)
	if err != nil {
		return 0, err
	}
	return peakBandwidth, nil
}

func GetAssetList(ctx context.Context, deviceIds []string, lang model.Language, option QueryOption) ([]*model.DeviceInfo, error) {
	var out []*model.DeviceInfo
	rawSql := fmt.Sprintf(`SELECT d.*, 
       IFNULL(l.continent, '') as continent, 
       IFNULL(l.country,'') as country, 
       IFNULL(l.province,'') as province, 
       IFNULL(l.city,'') as city FROM %s d left join %s l ON d.external_ip COLLATE utf8mb4_unicode_ci = l.ip WHERE device_id IN (?) ORDER by d.device_status_code`,
		tableNameDeviceInfo, fmt.Sprintf("%s_%s", tableNameLocation, lang))

	if option.Page > 0 && option.PageSize > 0 {
		offset := (option.Page - 1) * option.PageSize
		limit := option.PageSize
		rawSql = fmt.Sprintf("%s limit %d offset %d", rawSql, limit, offset)
	}

	query, args, err := sqlx.In(rawSql, deviceIds)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)
	err = DB.SelectContext(ctx, &out, query, args...)
	if err != nil {
		return nil, err
	}

	return out, nil
}
