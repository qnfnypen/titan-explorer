package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
)

// AssetGroup user asset group
type AssetGroup struct {
	ID          int64     `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Parent      int64     `db:"parent"`
	AssetCount  int64     `db:"asset_count"`
	AssetSize   int64     `db:"asset_size"`
	CreatedTime time.Time `db:"created_time"`
	ShareStatus int64     `db:"share_status"`
	VistitCount int64     `db:"visit_count"`
}

// OnlyUserGroupAsset 唯一存在的用户组文件
type OnlyUserGroupAsset struct {
	CID    string `db:"cid"`
	Num    int64  `db:"num"`
	AreaID string `db:"area_id"`
}

// ListAssetGroupRsp list  asset group records
type ListAssetGroupRsp struct {
	Total       int64         `json:"total"`
	AssetGroups []*AssetGroup `json:"infos"`
}

// GetUserAssetCountByGroupID Get count by group id
func getUserAssetCountByGroupID(ctx context.Context, uid string, groupID int) (int64, error) {
	var total int64
	query, args, err := squirrel.Select("COUNT(hash)").From(tableUserAsset).Where("user_id = ? AND group_id = ?", uid, groupID).ToSql()
	if err != nil {
		return 0, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, fmt.Errorf("get total of asset error:%w", err)
	}

	return total, nil
}

// DeleteAssetGroup delete asset group
func deleteAssetGroup(ctx context.Context, userID string, gid int) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Errorf("DeleteFileGroup Rollback err:%s", err.Error())
		}
	}()

	query := fmt.Sprintf(`DELETE FROM %s WHERE user_id=? AND parent=?`, tableNameAssetGroup)
	_, err = tx.ExecContext(ctx, query, userID, gid)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`DELETE FROM %s WHERE user_id=? AND id=?`, tableNameAssetGroup)
	_, err = tx.ExecContext(ctx, query, userID, gid)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// assetGroupExists is group exists
func assetGroupExists(ctx context.Context, uid string, gid int) (bool, error) {
	var id int64
	query, args, err := squirrel.Select("id").From(tableNameAssetGroup).Where("user_id = ? AND id = ?", uid, gid).ToSql()
	if err != nil {
		return false, fmt.Errorf("generate get asset's group sql error:%w", err)
	}
	err = DB.GetContext(ctx, &id, query, args...)
	if err != nil {
		return false, err
	}

	return id > 0, nil
}

func getAssetGroupParent(ctx context.Context, gid int) (int, error) {
	var parent int

	query := fmt.Sprintf("SELECT parent FROM %s WHERE id=?", tableNameAssetGroup)
	err := DB.GetContext(ctx, &parent, query, gid)
	if err != nil {
		return 0, err
	}

	return parent, nil
}
