package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/go-redis/redis/v9"
)

var tableNameLocation = "location"

var (
	GEOLocationKeyPrefix = "TITAN::GEO"
)

func CacheIPLocation(ctx context.Context, location *model.Location, lang model.Language) error {
	key := fmt.Sprintf("%s::%s::%s", GEOLocationKeyPrefix, lang, location.Ip)
	bytes, err := json.Marshal(location)
	if err != nil {
		return err
	}
	_, err = RedisCache.Set(ctx, key, bytes, 0).Result()
	return err
}

func GetCacheLocation(ctx context.Context, ip string, lang model.Language) (*model.Location, error) {
	key := fmt.Sprintf("%s::%s::%s", GEOLocationKeyPrefix, lang, ip)
	out := &model.Location{}
	bytes, err := RedisCache.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetLocationInfoByIp(ctx context.Context, ip string, out *model.Location, lang model.Language) error {
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE ip = ?`, fmt.Sprintf("%s_%s", tableNameLocation, lang)), ip,
	).StructScan(out); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func UpsertLocationInfo(ctx context.Context, out *model.Location, lang model.Language) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (ip, continent, country, province, city, longitude, latitude,area_code, isp, 
                zip_code, elevation,  created_at) 
		VALUES (:ip, :continent, :country, :province, :city, :longitude, :latitude, :area_code, :isp, :zip_code,
		 :elevation, :created_at) 
		 ON DUPLICATE KEY UPDATE continent = VALUES(continent), country = VALUES(country), province = VALUES(province), city = VALUES(city),
		longitude = VALUES(longitude), latitude = VALUES(latitude), area_code = VALUES(area_code), isp = VALUES(isp),
		zip_code = VALUES(zip_code), elevation = VALUES(elevation)`, fmt.Sprintf("%s_%s", tableNameLocation, lang)),
		out)
	return err
}
