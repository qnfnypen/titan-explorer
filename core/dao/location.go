package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameLocation = "location"

func GetLocationInfoByIp(ctx context.Context, ip string, out *model.Location) error {
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE ip = ?`, tableNameLocation), ip,
	).StructScan(out); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func UpsertLocationInfo(ctx context.Context, out *model.Location) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (ip, continent, country, province, city, longitude, latitude,area_code, isp, 
                zip_code, elevation,  created_at) 
		VALUES (:ip, :continent, :country, :province, :city, :longitude, :latitude, :area_code, :isp, :zip_code,
		 :elevation, :created_at) 
		 ON DUPLICATE KEY UPDATE continent = VALUES(continent), country = VALUES(country), province = VALUES(province), city = VALUES(city),
		longitude = VALUES(longitude), latitude = VALUES(latitude), area_code = VALUES(area_code), isp = VALUES(isp),
		zip_code = VALUES(zip_code), elevation = VALUES(elevation)`, tableNameLocation),
		out)
	return err
}
