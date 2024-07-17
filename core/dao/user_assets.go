package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableNameAssetGroup = "user_asset_group"
	tableUserAsset      = "user_asset"
	tableUserAssetVisit = "asset_visit_count"
)

type (
	// UserAssetDetail 用户表详情
	UserAssetDetail struct {
		UserID      string    `db:"user_id"`
		Hash        string    `db:"hash"`
		AreaID      string    `db:"area_id"`
		AssetName   string    `db:"asset_name"`
		AssetType   string    `db:"asset_type"`
		ShareStatus int64     `db:"share_status"`
		Expiration  time.Time `db:"expiration"`
		CreatedTime time.Time `db:"created_time"`
		TotalSize   int64     `db:"total_size"`
		Password    string    `db:"password"`
		GroupID     int64     `db:"group_id"`
		VisitCount  int64     `db:"visit_count"`
	}
)

// AddAssetAndUpdateSize 添加文件信息并修改使用的storage存储空间
func AddAssetAndUpdateSize(ctx context.Context, asset *model.UserAsset) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 添加文件记录
	query, args, err := squirrel.Insert(tableUserAsset).Columns("user_id,asset_name,asset_type,total_size,group_id,area_id,hash,created_time,expiration").
		Values(asset.UserID, asset.AssetName, asset.AssetType, asset.TotalSize, asset.GroupID, asset.AreaID, asset.Hash, asset.CreatedTime, asset.Expiration).ToSql()
	if err != nil {
		return fmt.Errorf("generate insert asset sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 修改用户storage已使用记录
	query, args, err = squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("used_storage_size + ?", asset.TotalSize)).Where("username = ?", asset.UserID).ToSql()
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
func DelAssetAndUpdateSize(ctx context.Context, hash, userID, areaID string, size int64) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// 删除文件记录
	query, args, err := squirrel.Delete(tableUserAsset).Where("hash = ? AND user_id = ? AND area_id = ?", hash, userID, areaID).ToSql()
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
func UpdateAssetShareStatus(ctx context.Context, hash, userID, areaID string) error {
	query, args, err := squirrel.Update(tableUserAsset).Set("share_status", 1).Where("hash = ? AND user_id = ? AND area_id = ?", hash, userID, areaID).ToSql()
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
func ListAssets(ctx context.Context, uid string, page, size, groupID int) (int64, []*UserAssetDetail, error) {
	var (
		total int64
		infos []*UserAssetDetail
	)

	total, err := getUserAssetCountByGroupID(ctx, uid, groupID)
	if err != nil {
		return 0, nil, err
	}

	query, args, err := squirrel.Select("ua.*,uav.count AS visit_count").From(fmt.Sprintf("%s AS ua", tableUserAsset)).LeftJoin("%s AS uav ON ua.hash=uav.hash", tableUserAssetVisit).
		Where("ua.user_id = ? AND ua.group_id = ?", uid, groupID).OrderBy("created_at desc").
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

	query, args, err = squirrel.Select("ag.*,COUNT(a.user_id) AS asset_count,COALESCE(SUM(a.total_size), 0) AS asset_size").From(fmt.Sprintf("%s as ag", tableNameAssetGroup)).
		LeftJoin(fmt.Sprintf("%s as a ON ag.user_id=a.user_id AND ag.id=a.group_id", tableUserAsset)).
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

// GetUserAsset 获取用户文件信息
func GetUserAsset(ctx context.Context, hash, uid, areaID string) (*UserAssetDetail, error) {
	var asset UserAssetDetail

	query, args, err := squirrel.Select("ua.*,uav.count AS visit_count").From(fmt.Sprintf("%s AS ua", tableUserAsset)).LeftJoin("%s AS uav ON ua.hash=uav.hash", tableUserAssetVisit).
		Where("ua.user_id = ? AND ua.hash = ? AND ua.area_id = ?", uid, hash, areaID).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &asset, query, args...)
	if err != nil {
		return nil, err
	}
	return &asset, err
}
