package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameAsset = "assets"
var tableNameProject = "projects"

func AddAssets(ctx context.Context, assets []*model.Asset) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s ( node_id, event, cid, hash, total_size, end_time, expiration, user_id, type, name, created_at, updated_at)
			VALUES ( :node_id, :event, :cid, :hash, :total_size, :end_time, :expiration, :user_id, :type, :name, :created_at, :updated_at) 
			ON DUPLICATE KEY UPDATE  event = VALUES(event), end_time = VALUES(end_time), expiration = VALUES(expiration), user_id = VALUES(user_id), type = VALUES(type), name = VALUES(name);`, tableNameAsset,
	), assets)
	return err
}

func UpdateAssetPath(ctx context.Context, cid string, path string) error {
	_, err := DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET path = ? where cid = ?`, tableNameAsset), path, cid)
	return err
}

func UpdateAssetEvent(ctx context.Context, cid string, event int) error {
	_, err := DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET event = ? where cid = ?`, tableNameAsset), event, cid)
	return err
}

func GetLatestAsset(ctx context.Context) (*model.Asset, error) {
	var asset model.Asset
	err := DB.GetContext(ctx, &asset, fmt.Sprintf(
		`SELECT * from %s ORDER BY created_at DESC LIMIT 1`, tableNameAsset))
	if err != nil {
		return nil, err
	}
	return &asset, err
}

func GetAssetsByEmptyPath(ctx context.Context) ([]*model.Asset, int64, error) {
	var out []*model.Asset
	err := DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s WHERE event = 1 AND path = ''`, tableNameAsset,
	))

	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s WHERE event = 1 AND path = ''`, tableNameAsset))
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetAssetByCID(ctx context.Context, cid string) (*model.Asset, error) {
	var asset model.Asset
	err := DB.GetContext(ctx, &asset, fmt.Sprintf(
		`SELECT * from %s WHERE cid = ?`, tableNameAsset), cid)
	if err != nil {
		return nil, err
	}
	return &asset, err
}

func CountAssets(ctx context.Context) ([]*model.StorageStats, error) {
	queryStatement := fmt.Sprintf(`select t.project_id, t.project_name, sum(t.total_size) as total_size, t.time, sum(t.gas) as gas, sum(t.pledge) as pledge, max(t.expiration) as expiration from (
		select a.project_id, s.name as project_name, sum(a.total_size) as total_size, DATE_FORMAT(now(),'%%Y-%%m-%%d %%H:%%i') as time, max(f.gas) as gas, max(f.pledge) as pledge, 
		max(f.end_time) as expiration  from %s a inner join %s s on a.project_id = s.id left join %s f on a.path = f.path 
		where a.path <> '' and a.event = 1  group by a.project_id, f.message_cid) t group by t.project_id`, tableNameAsset, tableNameProject, tableNameFilStorage)

	var out []*model.StorageStats
	if err := DB.SelectContext(ctx, &out, queryStatement); err != nil {
		return nil, err
	}

	userProviderInProject, err := getUserProviderInProject(ctx)
	if err != nil {
		return nil, err
	}

	for i, item := range out {
		uip, ok := userProviderInProject[item.ProjectId]
		if !ok {
			continue
		}
		out[i].UserCount = uip.UserCount
		out[i].ProviderCount = uip.ProviderCount
	}

	providerInProject, err := getProviderLocationInProject(ctx)
	if err != nil {
		return nil, err
	}

	for i, item := range out {
		uip, ok := providerInProject[item.ProjectId]
		if !ok {
			continue
		}
		out[i].Locations = uip.Locations
	}

	return out, nil
}

func getProviderLocationInProject(ctx context.Context) (map[int64]*model.StorageStats, error) {
	out := make(map[int64]*model.StorageStats)
	queryStatement := fmt.Sprintf(`select a.project_id, sp.location as locations from %s a left join %s f on a.path = f.path  
    left join %s sp on f.provider = sp.provider_id where a.path <> '' group by a.project_id, locations`, tableNameAsset, tableNameFilStorage, tableNameStorageProvider)

	rows, err := DB.Queryx(queryStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id       int64
			location sql.NullString
		)
		if err := rows.Scan(&id, &location); err != nil {
			return nil, err
		}
		if _, ok := out[id]; !ok {
			out[id] = &model.StorageStats{
				ProjectId: id,
				Locations: location.String,
			}
			continue
		}

		if out[id].Locations == "" {
			out[id].Locations = location.String
		} else {
			out[id].Locations += "," + location.String
		}
	}

	return out, nil
}

func getUserProviderInProject(ctx context.Context) (map[int64]*model.StorageStats, error) {
	out := make(map[int64]*model.StorageStats)
	queryStatement := fmt.Sprintf(`select a.project_id, count(DISTINCT a.user_id) as user_count, count(DISTINCT f.provider) as provider_count from %s a 
    left join %s f on a.path = f.path where a.path <> '' group by a.project_id`, tableNameAsset, tableNameFilStorage)

	rows, err := DB.Queryx(queryStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.StorageStats
		if err := rows.StructScan(&s); err != nil {
			return nil, err
		}
		out[s.ProjectId] = &s
	}

	return out, nil
}
