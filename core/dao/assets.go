package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableNameAsset      = "assets"
	tableNameProject    = "projects"
	tableNameAssetGroup = "asset_group"
)

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
		`SELECT * from %s ORDER BY end_time DESC LIMIT 1`, tableNameAsset))
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
			if location.String != "" {
				out[id].Locations += "," + location.String
			}
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

// AddAssetAndUpdateSize 添加文件信息并修改使用的storage存储空间
func AddAssetAndUpdateSize(ctx context.Context, asset *model.Asset) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 添加文件记录
	query, args, err := squirrel.Insert(tableNameAsset).Columns("user_id,name,cid,type,node_id,total_size,group_id,area_id,hash,event,project_id").
		Values(asset.UserId, asset.Name, asset.Cid, asset.Type, asset.NodeID, asset.TotalSize, asset.GroupID, asset.AreaID, asset.Hash, asset.Event, asset.ProjectId).ToSql()
	if err != nil {
		return fmt.Errorf("generate insert asset sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 修改用户storage已使用记录
	query, args, err = squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("used_storage_size + ?", asset.TotalSize)).Where("username = ?", asset.UserId).ToSql()
	if err != nil {
		return fmt.Errorf("generate update users sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DelAssetAndUpdateSize 删除文件信息并修改使用的storage存储空间
func DelAssetAndUpdateSize(ctx context.Context, cid, userID string, size int64) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// 删除文件记录
	query, args, err := squirrel.Delete(tableNameAsset).Where("cid = ? AND user_id = ?", cid, userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 修改用户storage已使用记录
	query, args, err = squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("used_storage_size - ?", size)).Where("username = ?", userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update users sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateAssetShareStatus 修改文件分享状态
func UpdateAssetShareStatus(ctx context.Context, cid, userID string) error {
	query, args, err := squirrel.Update(tableNameAsset).Set("share_status", 1).Where("user_id = ? AND cid = ?", userID, cid).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// ListAssets 获取对应文件夹的文件列表
func ListAssets(ctx context.Context, uid string, page, size, groupID int) (int64, []*model.Asset, error) {
	var (
		total int64
		infos []*model.Asset
	)

	total, err := getUserAssetCountByGroupID(ctx, uid, groupID)
	if err != nil {
		return 0, nil, err
	}

	query, args, err := squirrel.Select("*").From(tableNameAsset).Where("user_id = ? AND group_id = ?", uid, groupID).OrderBy("created_at desc").
		Limit(uint64(size)).Offset(uint64((page - 1) * size)).ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &infos, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get list of asset error:%w", err)
	}

	return total, infos, nil
}

// CreateAssetGroup 创建文件夹
func CreateAssetGroup(ctx context.Context, uid, name string, parent int) (*AssetGroup, error) {
	var id int64

	// 判断是否是根目录
	if parent != 0 {
		exists, err := assetGroupExists(ctx, uid, parent)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, &api.ErrWeb{Code: terrors.GroupNotExist.Int(), Message: fmt.Sprintf("CreateAssetGroup failed, group parent [%d] is not exist ", parent)}
		}
	}

	// 获取数量
	query, args, err := squirrel.Select("count(id)").From(tableNameAssetGroup).Where("user_id = ?", uid).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset's group sql error:%w", err)
	}
	err = DB.GetContext(ctx, &id, query, args...)
	if err != nil {
		return nil, err
	}
	if id >= 20 {
		return nil, &api.ErrWeb{Code: terrors.GroupLimit.Int(), Message: fmt.Sprintf("CreateAssetGroup failed, Exceed the limit %d", 20)}
	}

	// 插入数据
	createdTime := time.Now()
	query, args, err = squirrel.Insert(tableNameAssetGroup).Columns("user_id,name,parent,created_time").Values(uid, name, parent, createdTime).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate insert asset's group sql error:%w", err)
	}
	res, err := DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	id, _ = res.LastInsertId()

	return &AssetGroup{ID: id, UserID: uid, Name: name, Parent: int64(parent), CreatedTime: createdTime}, nil
}

// ListAssetGroupForUser 根据用户获取对应的文件夹信息
func ListAssetGroupForUser(ctx context.Context, uid string, parent, limit, offset int) (*ListAssetGroupRsp, error) {
	resp := new(ListAssetGroupRsp)
	resp.AssetGroups = make([]*AssetGroup, 0)

	query, args, err := squirrel.Select("count(id)").From(tableNameAssetGroup).Where("user_id=? AND parent=?", uid, parent).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset's group sql error:%w", err)
	}
	err = DB.GetContext(ctx, &resp.Total, query, args...)
	if err != nil {
		return nil, err
	}

	query, args, err = squirrel.Select("ag.*,COUNT(a.user_id) AS asset_count,COALESCE(SUM(a.total_size), 0) AS asset_size").From(fmt.Sprintf("%s as ag", tableNameAssetGroup)).LeftJoin(fmt.Sprintf("%s as a ON ag.user_id=a.user_id AND ag.id=a.group_id", tableNameAsset)).
		Where("ag.user_id=? AND ag.parent=?", uid, parent).GroupBy("ag.id").OrderBy("ag.created_time DESC").Limit(uint64(limit)).Offset(uint64(offset)).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset's group sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &resp.AssetGroups, query, args...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteAssetGroup delete asset group
func DeleteAssetGroup(ctx context.Context, uid string, gid int) error {
	gCount, err := getUserAssetCountByGroupID(ctx, uid, gid)
	if err != nil {
		return err
	}

	if gCount > 0 {
		return &api.ErrWeb{Code: terrors.GroupNotEmptyCannotBeDelete.Int(), Message: "There are assets in the group and the group cannot be deleted"}
	}

	rsp, err := ListAssetGroupForUser(ctx, uid, gid, 1, 0)
	if err != nil {
		return err
	}
	if rsp.Total > 0 {
		return &api.ErrWeb{Code: terrors.GroupNotEmptyCannotBeDelete.Int(), Message: "There are assets in the group and the group cannot be deleted"}
	}

	return deleteAssetGroup(ctx, uid, gid)
}

// UpdateAssetGroupName update user asset group name
func UpdateAssetGroupName(ctx context.Context, uid, rename string, groupID int) error {
	query, args, err := squirrel.Update(tableNameAssetGroup).Set("name", rename).Where("user_id = ? AND id = ?", uid, groupID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset's group sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)

	return err
}

// MoveAssetGroup move a asset group
func MoveAssetGroup(ctx context.Context, userID string, groupID, targetGroupID int) error {
	if groupID == 0 {
		return &api.ErrWeb{Code: terrors.RootGroupCannotMoved.Int(), Message: "the root group cannot be moved"}
	}

	if groupID == targetGroupID {
		return &api.ErrWeb{Code: terrors.GroupsAreSame.Int(), Message: "groups are the same"}
	}

	if targetGroupID != 0 {
		exists, err := assetGroupExists(ctx, userID, targetGroupID)
		if err != nil {
			return err
		}
		if !exists {
			return &api.ErrWeb{Code: terrors.GroupNotExist.Int(), Message: fmt.Sprintf("CreateAssetGroup failed, group parent [%d] is not exist ", targetGroupID)}
		}

		// Prevent loops
		gid := targetGroupID
		for {
			gid, err = getAssetGroupParent(ctx, gid)
			if err != nil {
				return err
			}

			if gid == groupID {
				return &api.ErrWeb{Code: terrors.CannotMoveToSubgroup.Int(), Message: "cannot move to subgroup"}
			}

			if gid == 0 {
				break
			}
		}
	}

	query, args, err := squirrel.Update(tableNameAssetGroup).Set("parent", targetGroupID).Where("user_id=? AND id=?", userID, groupID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset's group sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)

	return err
}

// UpdateAssetGroup update user asset group
func UpdateAssetGroup(ctx context.Context, userID, cid string, groupID int) error {
	query, args, err := squirrel.Update(tableNameAsset).Set("group_id", groupID).Where("user_id=? AND cid=?", userID, cid).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)

	return err
}

func AllAssets(ctx context.Context) ([]*model.Asset, error) {
	var assets []*model.Asset
	err := DB.SelectContext(ctx, &assets, fmt.Sprintf("SELECT * FROM %s", tableNameAsset))
	if err != nil {
		return nil, err
	}
	return assets, nil
}
