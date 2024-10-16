package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableNameAssetGroup   = "user_asset_group"
	tableUserAsset        = "user_asset"
	tableUserAssetVisit   = "asset_visit_count"
	tableUserAssetArea    = "user_asset_area"
	tableTempAsset        = "temp_asset"
	tableAssetStorageHour = "asset_storage_hour"
	tableUserAssetMap     = "user_asset_map"
)

type (
	// KVMap 键值映射
	KVMap struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	// UserAssetDetail 用户表详情
	UserAssetDetail struct {
		UserID      string    `db:"user_id"`
		Hash        string    `db:"hash"`
		Cid         string    `db:"cid"`
		AreaIDs     []string  `db:"-" json:"area_ids"`
		AreaIDMaps  []KVMap   `db:"-" json:"area_maps"`
		AssetName   string    `db:"asset_name"`
		AssetType   string    `db:"asset_type"`
		ShareStatus int64     `db:"share_status"`
		Expiration  time.Time `db:"expiration"`
		CreatedTime time.Time `db:"created_time"`
		TotalSize   int64     `db:"total_size"`
		Password    string    `db:"password"`
		GroupID     int64     `db:"group_id"`
		VisitCount  int64     `db:"visit_count"`
		// ShortPass   string    `db:"short_pass"`
		// IsSync      bool      `db:"is_sync" json:"-"`
	}
	// SubAssetDetail 部分用户文件信息
	SubAssetDetail struct {
		Cid       string `db:"cid"`
		TotalSize int64  `db:"total_size"`
	}

	// DashBoardInfo 仪表盘数据信息
	DashBoardInfo struct {
		Date           int64  `db:"hour" json:"-"`
		DateStr        string `db:"-" json:"date"`
		DownloadCount  int64  `db:"download_count" json:"DownloadCount"`
		PeakBandwidth  int64  `db:"peak_bandwidth" json:"PeakBandwidth"`
		TotalBandwidth int64  `db:"total_traffic" json:"TotalBandwidth"`
	}
	// UserStorageFlowInfo 用户存储流量信息
	UserStorageFlowInfo struct {
		TotalTraffic  int64 `db:"total_traffic"`
		PeakBandwidth int64 `db:"peak_bandwidth"`
	}
)

// AddAssetAndUpdateSize 添加文件信息并修改使用的storage存储空间
func AddAssetAndUpdateSize(ctx context.Context, asset *model.UserAsset, areaIDs []string, syncArea string) error {
	tx, err := DB.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	if len(areaIDs) == 0 {
		return errors.New("area id can not be empty")
	}

	// 查询文件记录是否存在
	ua, err := GetUserAsset(ctx, asset.Hash, asset.UserID)
	if err != nil && err != sql.ErrNoRows {
		log.Error(err)
		return fmt.Errorf("generate insert asset sql error:%w", err)
	}
	if err == sql.ErrNoRows {
		// 添加文件记录，判断文件是否存在，不存在则新增
		query, args, err := squirrel.Insert(tableUserAsset).Columns("user_id,asset_name,asset_type,total_size,group_id,hash,created_time,expiration,password,cid,md5,extra_id").
			Values(asset.UserID, asset.AssetName, asset.AssetType, asset.TotalSize, asset.GroupID, asset.Hash, asset.CreatedTime, asset.Expiration, asset.Password, asset.Cid, asset.MD5, asset.ExtraID).ToSql()
		if err != nil {
			log.Error(err)
			return fmt.Errorf("generate insert asset sql error:%w", err)
		}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}
	// 添加文件区域,只有第一个为插入时才更新，后续不变
	query, args, err := squirrel.Insert(tableUserAssetArea).Columns("hash,user_id,area_id,is_sync").Values(asset.Hash, asset.UserID, syncArea, true).Suffix("ON DUPLICATE KEY UPDATE is_sync = VALUES(is_sync)").ToSql()
	if err != nil {
		log.Error(err)
		return fmt.Errorf("generate insert asset's area sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error(err)
		return err
	}
	// 后续不变，notExist继续插入或保持原来不变
	if len(areaIDs) > 0 {
		abuiler := squirrel.Insert(tableUserAssetArea).Columns("hash,user_id,area_id,is_sync")
		for _, v := range areaIDs {
			isSync := false
			abuiler = abuiler.Values(asset.Hash, asset.UserID, v, isSync)
		}
		query, args, err = abuiler.Options("IGNORE").ToSql()
		if err != nil {
			log.Error(err)
			return fmt.Errorf("generate insert asset's area sql error:%w", err)
		}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	// 增加用户文件映射关系
	query, args, err = squirrel.Insert(tableUserAssetMap).Columns("user_id", "asset_hash").Values(asset.UserID, asset.Hash).Options("IGNORE").ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of insert user_asset_map error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert user_asset_map error:%w", err)
	}
	// 修改用户storage已使用记录
	if ua == nil || ua.UserID == "" {
		query, args, err = squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("used_storage_size + ?", asset.TotalSize)).Where("username = ?", asset.UserID).ToSql()
		if err != nil {
			log.Error(err)
			return fmt.Errorf("generate update users sql error:%w", err)
		}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return tx.Commit()
}

// DelAssetAndUpdateSize 删除文件信息并修改使用的storage存储空间
func DelAssetAndUpdateSize(ctx context.Context, hash, userID string, areaID []string, isNeedDel bool) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// 删除文件记录
	query, args, err := squirrel.Delete(tableUserAssetArea).Where(squirrel.Eq{
		"area_id": areaID,
		"user_id": userID,
		"hash":    hash,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 判断是否需要删除文件记录
	if !isNeedDel {
		return tx.Commit()
	}
	// 获取文件尺寸大小
	var sa SubAssetDetail
	query, args, err = squirrel.Select("total_size,cid").From(tableUserAsset).Where("hash = ? AND user_id = ?", hash, userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql error:%w", err)
	}
	err = tx.GetContext(ctx, &sa, query, args...)
	if err != nil {
		return err
	}
	// 删除文件记录
	query, args, err = squirrel.Delete(tableUserAsset).Where("hash = ? AND user_id = ?", hash, userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 修改用户storage已使用记录
	query, args, err = squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("GREATEST(used_storage_size - ?,0)", sa.TotalSize)).Where("username = ?", userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update users sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// 删除该文件的分享
	query, args, err = squirrel.Delete(tableNameLink).Where("username = ?", userID).Where("cid = ?", sa.Cid).ToSql()
	if err != nil {
		return fmt.Errorf("generate delete links sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// 删除下载次数记录
	query, args, err = squirrel.Delete(tableUserAssetVisit).Where("user_id = ? AND hash = ?", userID, hash).ToSql()
	if err != nil {
		return fmt.Errorf("generate delete assest_visit_count sql error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateAssetShareStatus 修改文件分享状态
func UpdateAssetShareStatus(ctx context.Context, hash, userID string) error {
	query, args, err := squirrel.Update(tableUserAsset).Set("share_status", 1).Where("hash = ? AND user_id = ?", hash, userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateGroupShareStatus 修改文件组分享状态
func UpdateGroupShareStatus(ctx context.Context, userID string, groupID int64) error {
	query, args, err := squirrel.Update(tableNameAssetGroup).Set("share_status", 1).Where("user_id = ? AND id = ?", userID, groupID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update user_asset_group sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// ListAssets 获取对应文件夹的文件列表
func ListAssets(ctx context.Context, uid string, limit, offset, groupID int) (int64, []*UserAssetDetail, error) {
	var (
		total int64
		infos []*UserAssetDetail
	)

	total, err := getUserAssetCountByGroupID(ctx, uid, groupID)
	if err != nil {
		return 0, nil, err
	}

	query, args, err := squirrel.Select("ua.user_id,ua.hash,ua.cid,ua.asset_name,ua.asset_type,ua.share_status,ua.expiration,ua.created_time,ua.total_size,ua.password,ua.group_id,IFNULL(uav.count,0) AS visit_count").
		From(fmt.Sprintf("%s AS ua", tableUserAsset)).LeftJoin(fmt.Sprintf("%s AS uav ON ua.hash=uav.hash and ua.user_id = uav.user_id", tableUserAssetVisit)).
		Where("ua.user_id = ? AND ua.group_id = ?", uid, groupID).OrderBy("ua.created_time desc").
		Limit(uint64(limit)).Offset(uint64(offset)).ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &infos, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get list of asset error:%w", err)
	}

	return total, infos, nil
}

// UpdateAssetCid 更新文件的cid信息
func UpdateAssetCid(ctx context.Context, hash, cid string) error {
	query, args, err := squirrel.Update(tableUserAsset).Set("cid", cid).Where("hash = ?", hash).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql to update cid of user_asset error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
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

// GetUserAssetGroupInfo
func GetUserAssetGroupInfo(ctx context.Context, uid string, gid int) (*AssetGroup, error) {
	var group AssetGroup
	query, args, err := squirrel.Select("ag.*,COUNT(a.user_id) AS asset_count,COALESCE(SUM(a.total_size), 0) AS asset_size").From(fmt.Sprintf("%s as ag", tableNameAssetGroup)).
		LeftJoin(fmt.Sprintf("%s as a ON ag.user_id=a.user_id AND ag.id=a.group_id", tableUserAsset)).
		Where("ag.user_id=? AND ag.id=?", uid, gid).Limit(1).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset's group sql error:%w", err)
	}
	err = DB.GetContext(ctx, &group, query, args...)
	if err != nil {
		return nil, err
	}

	return &group, nil
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

// UpdateAssetName 更新用户文件名
func UpdateAssetName(ctx context.Context, newName, uid, hash string) error {
	query, args, err := squirrel.Update(tableUserAsset).Set("asset_name", newName).Where("user_id = ? AND hash = ?", uid, hash).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset's sql error:%w", err)
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
func UpdateAssetGroup(ctx context.Context, userID, hash string, groupID int) error {
	query, args, err := squirrel.Update(tableUserAsset).Set("group_id", groupID).Where("user_id=? AND hash=?", userID, hash).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)

	return err
}

// GetUserAssetDetail 获取用户文件信息
func GetUserAssetDetail(ctx context.Context, hash, uid string) (*UserAssetDetail, error) {
	var asset UserAssetDetail

	query, args, err := squirrel.Select("ua.user_id,ua.hash,ua.asset_name,ua.asset_type,ua.share_status,ua.expiration,ua.created_time,ua.total_size,ua.password,ua.group_id,IFNULL(uav.count,0) AS visit_count").
		From(fmt.Sprintf("%s AS ua", tableUserAsset)).LeftJoin(fmt.Sprintf("%s AS uav ON ua.hash=uav.hash and ua.user_id = uav.user_id", tableUserAssetVisit)).
		Where("ua.user_id = ? AND ua.hash = ?", uid, hash).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &asset, query, args...)
	if err != nil {
		return nil, err
	}
	return &asset, err
}

// GetUserAsset 获取用户文件信息
func GetUserAsset(ctx context.Context, hash, uid string) (*model.UserAsset, error) {
	var out model.UserAsset
	query := "SELECT * FROM user_asset where hash = ? and user_id = ?"
	err := DB.GetContext(ctx, &out, query, hash, uid)
	return &out, err
}

func GetUserAssetByBuilder(ctx context.Context, sb squirrel.SelectBuilder) (*model.UserAsset, error) {
	var out model.UserAsset
	query, args, err := sb.From(tableUserAsset).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	err = DB.SelectContext(ctx, &out, query, args...)
	return &out, err
}

func UpdateUserAsset(ctx context.Context, asset *model.UserAsset) error {
	query, args, err := squirrel.Update(tableUserAsset).SetMap(map[string]interface{}{
		"asset_name":   asset.AssetName,
		"asset_type":   asset.AssetType,
		"share_status": asset.ShareStatus,
		"expiration":   asset.Expiration,
		"total_size":   asset.TotalSize,
		"password":     asset.Password,
		// "short_pass":   asset.ShortPass,
		"group_id": asset.GroupID,
	}).Where("hash = ? AND user_id = ?", asset.Hash, asset.UserID).ToSql()
	if err != nil {
		return fmt.Errorf("generate update asset sql error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args...)

	return err
}

// GetUserAssetNotAreaIDs 返回不存在的area_id
func GetUserAssetNotAreaIDs(ctx context.Context, hash, uid string, areaID []string) ([]string, error) {
	var (
		aids, notExistAids []string
	)

	query, args, err := squirrel.Select("area_id").From(tableUserAssetArea).Where(squirrel.Eq{
		"area_id": areaID,
		"user_id": uid,
		"hash":    hash,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}

	err = DB.SelectContext(ctx, &aids, query, args...)
	if err != nil {
		return nil, err
	}

	notExistAids = getNotExists(areaID, aids)

	return notExistAids, nil
}

// CheckAssetIsSyncByAreaID 判断
func CheckAssetIsSyncByAreaID(ctx context.Context, hash, areaID string) bool {
	var isSync bool

	query, args, err := squirrel.Select("is_sync").From(tableUserAssetArea).Where("area_id = ? AND hash = ?", areaID, hash).ToSql()
	if err != nil {
		log.Errorf("generate get asset sql error:%v", err)
		return false
	}
	err = DB.GetContext(ctx, &isSync, query, args...)
	if err != nil {
		log.Errorf("get asset sql error:%v", err)
		return false
	}

	return isSync
}

func getNotExists(pAids, nAids []string) []string {
	var (
		notExistAids []string
		aidMaps      = make(map[string]int)
	)

	for _, v := range pAids {
		aidMaps[v] = 1
	}
	for _, v := range nAids {
		if _, ok := aidMaps[v]; ok {
			delete(aidMaps, v)
		}
	}
	for k := range aidMaps {
		notExistAids = append(notExistAids, k)
	}

	return notExistAids
}

// GetUnSyncAreaIDs 获取未同步的区域
func GetUnSyncAreaIDs(ctx context.Context, uid, hash string) ([]string, error) {
	var aids []string

	sq := squirrel.Select("DISTINCT(area_id)").From(tableUserAssetArea).Where("hash = ? AND is_sync = 0", hash)
	if uid != "" {
		sq = sq.Where("user_id = ?", uid)
	}
	query, args, err := sq.ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}

	err = DB.SelectContext(ctx, &aids, query, args...)
	if err != nil {
		return nil, err
	}

	return aids, nil
}

// UpdateUnSyncAreaIDs 更新未同步的区域
func UpdateUnSyncAreaIDs(ctx context.Context, uid, hash string, aids []string) error {
	query, args, err := squirrel.Update(tableUserAssetArea).Set("is_sync", true).Where(squirrel.Eq{
		"area_id": aids,
		"user_id": uid,
		"hash":    hash,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate get asset sql error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

// GetOneSyncSuccessArea 随机获取一个同步完成的区域
func GetOneSyncSuccessArea(ctx context.Context, hash string) (string, error) {
	var aids []string

	query, args, err := squirrel.Select("DISTINCT(area_id)").From(tableUserAssetArea).Where("hash = ? AND is_sync = 1", hash).ToSql()
	if err != nil {
		return "", fmt.Errorf("generate get asset sql error:%w", err)
	}

	err = DB.SelectContext(ctx, &aids, query, args...)
	if err != nil {
		return "", err
	}
	if len(aids) == 0 {
		return "", errors.New("now rows")
	}

	return aids[rand.Intn(len(aids))], nil
}

// CheckAssetHashIsExist 判断文件hash是否存在
func CheckAssetHashIsExist(ctx context.Context, hash string) bool {
	var count int64

	query, args, err := squirrel.Select("COUNT(area_id)").From(tableUserAssetArea).Where("hash = ? AND is_sync = 0", hash).ToSql()
	if err != nil {
		return false
	}

	err = DB.SelectContext(ctx, &count, query, args...)
	if err != nil {
		return false
	}

	return count > 0
}

func UpdateSyncAssetAreas(ctx context.Context, areaID string, hashs []string) error {
	query, args, err := squirrel.Update(tableUserAssetArea).Set("is_sync", true).Where(squirrel.Eq{
		"area_id": areaID,
		"hash":    hashs,
		"is_sync": 0,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate get asset sql error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

// AddVisitCount 增加文件访问次数
func AddVisitCount(ctx context.Context, hash string, user_id string) error {
	if hash == "" {
		return nil
	}
	query, args, err := squirrel.Insert(tableUserAssetVisit).Columns("hash", "count", "user_id").Values(hash, 1, user_id).Suffix("ON DUPLICATE KEY UPDATE count = count + 1").ToSql()
	if err != nil {
		return fmt.Errorf("generate asset's visit count sql error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

func GetVisitCount(ctx context.Context, hash, user_id string) (int64, error) {
	var count int64

	query, args, err := squirrel.Select("count").From(tableUserAssetVisit).Where("hash = ? AND user_id = ?", hash, user_id).ToSql()
	if err != nil {
		return 0, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &count, query, args...)

	return count, err
}

// CheckUserAseetNeedDel 判断文件是否需要删除
func CheckUserAseetNeedDel(ctx context.Context, hash, uid string, areaID []string) ([]string, bool, error) {
	var (
		aids, existAids []string
	)

	// 获取文件所有调度器区域
	query, args, err := squirrel.Select("area_id").From(tableUserAssetArea).Where("user_id = ? AND hash = ?", uid, hash).ToSql()
	if err != nil {
		return nil, false, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &aids, query, args...)
	if err != nil {
		return nil, false, err
	}
	if len(areaID) == 0 {
		return aids, true, nil
	}
	// 获取指定区域中存在的调度器区域
	query, args, err = squirrel.Select("area_id").From(tableUserAssetArea).Where(squirrel.Eq{
		"area_id": areaID,
		"user_id": uid,
		"hash":    hash,
	}).ToSql()
	if err != nil {
		return nil, false, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &existAids, query, args...)
	if err != nil {
		return nil, false, err
	}

	return existAids, len(aids) == len(existAids), nil
}

// GetUserAssetAreaIDs 获取用户文件的调度器区域
func GetUserAssetAreaIDs(ctx context.Context, hash, uid string) ([]string, error) {
	var aids []string

	query, args, err := squirrel.Select("area_id").From(tableUserAssetArea).Where("user_id = ? AND hash = ? AND is_sync = 1", uid, hash).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &aids, query, args...)
	if err != nil {
		return nil, err
	}

	return aids, nil
}

// CheckUserAssetIsOnly 判断用户文件是否为唯一存在的
func CheckUserAssetIsOnly(ctx context.Context, hash, areaID string) (bool, error) {
	var num int64

	query, args, err := squirrel.Select("COUNT(hash)").From(tableUserAssetArea).Where("hash = ? AND area_id = ?", hash, areaID).ToSql()
	if err != nil {
		return false, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &num, query, args...)
	if err != nil {
		return false, err
	}

	return num <= 1, nil
}

// CheckUserAssetIsInAreaID 判断用户文件是否存在于指定区域
func CheckUserAssetIsInAreaID(ctx context.Context, userID, hash, areaID string) (bool, error) {
	var num int64

	query, args, err := squirrel.Select("COUNT(hash)").From(tableUserAssetArea).Where("user_id = ? AND hash = ? AND area_id LIKE ? AND is_sync = 1", userID, hash, `%`+areaID+`%`).ToSql()
	if err != nil {
		return false, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &num, query, args...)
	if err != nil {
		return false, err
	}

	return num >= 1, nil
}

// GetOneAreaIDByAreaID 根据给定的areaID模糊获取一个准确的areaid
func GetOneAreaIDByAreaID(ctx context.Context, userID, hash, areaID string) (string, error) {
	var areaIDs []string

	query, args, err := squirrel.Select("area_id").From(tableUserAssetArea).Where("user_id = ? AND hash = ? AND area_id LIKE ? AND is_sync = 1", userID, hash, `%`+areaID+`%`).ToSql()
	if err != nil {
		return "", fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.SelectContext(ctx, &areaIDs, query, args...)
	if err != nil {
		return "", err
	}

	return areaIDs[0], nil
}

// GetTempAssetInfo 获取临时文件的信息
func GetTempAssetInfo(ctx context.Context, hash string) (*model.TempAsset, error) {
	var info model.TempAsset

	query, args, err := squirrel.Select("*").From(tableTempAsset).Where("hash = ?", hash).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get asset sql error:%w", err)
	}
	err = DB.GetContext(ctx, &info, query, args...)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// AddTempAssetShareCount 增加临时文件的分享信息
func AddTempAssetShareCount(ctx context.Context, hash string) error {
	if hash == "" {
		return nil
	}
	query, args, err := squirrel.Insert(tableTempAsset).Columns("hash").Values(hash).Suffix("ON DUPLICATE KEY UPDATE share_count = share_count + 1").ToSql()
	if err != nil {
		return fmt.Errorf("generate asset's temp asset sql error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

// AddTempAssetDownloadCount 增加临时文件的下载次数
func AddTempAssetDownloadCount(ctx context.Context, hash string) error {
	if hash == "" {
		return nil
	}

	query, args, err := squirrel.Update(tableTempAsset).Set("download_count", squirrel.Expr("download_count + ?", 1)).Where("hash = ?", hash).ToSql()
	if err != nil {
		return fmt.Errorf("generate asset's temp asset sql error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

// GetAreaIDsByHash 通过hash获取areaids
func GetAreaIDsByHash(ctx context.Context, hash string) ([]string, error) {
	var areaIDs []string

	query, args, err := squirrel.Select("DISTINCT(area_id)").From(tableUserAssetArea).Where("hash = ?", hash).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate asset's areaid sql error:%w", err)
	}

	err = DB.SelectContext(ctx, &areaIDs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get asset's areaids error:%w", err)
	}

	return areaIDs, nil
}

// AddAssetHourStorages 批量存储文件小时数据
func AddAssetHourStorages(ctx context.Context, ahss []model.AssetStorageHour) error {
	sq := squirrel.Insert(tableAssetStorageHour).Columns("hash,user_id,total_traffic,peak_bandwidth,download_count,timestamp")

	for _, v := range ahss {
		sq = sq.Values(v.Hash, v.UserID, v.TotalTraffic, v.PeakBandwidth, v.DownloadCount, v.TimeStamp)
	}

	query, args, err := sq.ToSql()
	if err != nil {
		return fmt.Errorf("generate insert asset_storage_hour error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	return err
}

// GetUserDashboardInfos 通过用户id获取最近24小时的信息
func GetUserDashboardInfos(ctx context.Context, uid string, ts time.Time) ([]DashBoardInfo, error) {
	var (
		list, nlist []DashBoardInfo
		timeMaps    = make(map[int64]DashBoardInfo)
	)

	st := ts.Add(-24 * time.Hour).Unix()

	// UNIX_TIMESTAMP(STR_TO_DATE(FROM_UNIXTIME(timestamp, '%Y-%m-%d %H:00:00'), '%Y-%m-%d %H:%i:%s'))
	query, args, err := squirrel.Select("UNIX_TIMESTAMP(STR_TO_DATE(FROM_UNIXTIME(timestamp-1, '%Y-%m-%d %H:00:00'), '%Y-%m-%d %H:%i:%s')) AS hour,SUM(download_count) AS download_count,max(peak_bandwidth) AS peak_bandwidth,SUM(total_traffic) AS total_traffic").
		From(tableAssetStorageHour).
		// Where(squirrel.Expr("hash IN (?)", squirrel.Select("hash").From(tableUserAsset).Where("user_id = ?", uid))).
		Where("timestamp < ? AND timestamp > ? AND user_id = ?", ts.Unix(), st, uid).GroupBy("hour").OrderBy("hour DESC").ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get user storage hour info error:%w", err)
	}

	err = DB.SelectContext(ctx, &list, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get user storage hour info error:%w", err)
	}

	for i, v := range list {
		v.Date += 3600
		list[i].DateStr = fmt.Sprintf("%d:00", time.Unix(v.Date, 0).Hour())
		timeMaps[v.Date] = list[i]
	}

	ht := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, ts.Location())
	for i := 23; i >= 0; i-- {
		tn := ht.Add(time.Hour * time.Duration(-i))
		if v, ok := timeMaps[tn.Unix()]; ok {
			nlist = append(nlist, v)
		} else {
			nlist = append(nlist, DashBoardInfo{DateStr: fmt.Sprintf("%d:00", tn.Hour())})
		}
	}

	return nlist, nil
}

// UserAssetAreaIDs 获取用户文件的区域id
func UserAssetAreaIDs(ctx context.Context, uid, hash string) ([]string, error) {
	var areaIds []string

	query, args, err := squirrel.Select("DISTINCT(area_id)").From(tableUserAssetArea).Where("hash = ? AND user_id = ?", hash, uid).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get area_id error:%w", err)
	}

	err = DB.SelectContext(ctx, &areaIds, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get area_id error:%w", err)
	}

	return areaIds, nil
}

// GetHashAreaIDList 根据用户信息获取区域hash列表
func GetHashAreaIDList(ctx context.Context, uid string) (map[string][]string, error) {
	var (
		areaInfos []model.UserAssetArea
		areaHashs = make(map[string][]string)
	)

	query, args, err := squirrel.Select("*").From(tableUserAssetArea).Where("user_id = ? AND is_sync = 1", uid).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get asset_areas error:%w", err)
	}

	err = DB.SelectContext(ctx, &areaInfos, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get asset_areas error:%w", err)
	}

	for _, v := range areaInfos {
		if _, ok := areaHashs[v.AreaID]; ok {
			areaHashs[v.AreaID] = append(areaHashs[v.AreaID], v.Hash)
		} else {
			areaHashs[v.AreaID] = []string{v.Hash}
		}
	}

	return areaHashs, nil
}

// GetUserStorageFlowInfo 获取用户存储流量信息
func GetUserStorageFlowInfo(ctx context.Context, uid string) (*UserStorageFlowInfo, error) {
	var info = new(UserStorageFlowInfo)

	query, args, err := squirrel.Select("IFNULL(SUM(total_traffic),0) AS total_traffic,IFNULL(MAX(peak_bandwidth),0) AS peak_bandwidth").From(tableAssetStorageHour).
		// Where(squirrel.Expr("hash IN (?)", squirrel.Select("hash").From(tableUserAsset).Where("user_id = ?", uid))).
		Where("timestamp < ? AND user_id = ?", time.Now().Unix(), uid).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get storage flow error:%w", err)
	}

	err = DB.GetContext(ctx, info, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get storage flow error:%w", err)
	}

	return info, nil
}

// DeleteAssetGroupAndUpdateSize 删除文件组并更新用户已使用空间大小
func DeleteAssetGroupAndUpdateSize(ctx context.Context, userID string, gid int) error {
	var (
		gids  = []int{gid}
		ids   = []int{gid}
		tsize int64
	)

	// 递归获取所有的文件组id
	for {
		if len(ids) == 0 {
			break
		}
		query, args, err := squirrel.Select("id").From(tableNameAssetGroup).Where(squirrel.Eq{
			"user_id": userID,
			"parent":  ids,
		}).ToSql()
		if err != nil {
			return fmt.Errorf("generate get ids error:%w", err)
		}
		err = DB.SelectContext(ctx, &ids, query, args...)
		if err != nil {
			return fmt.Errorf("get ids error:%w", err)
		}
		gids = append(gids, ids...)
	}
	// 获取要删除的所有文件大小
	query, args, err := squirrel.Select("IFNULL(SUM(total_size),0)").From(tableUserAsset).Where(squirrel.Eq{
		"user_id":  userID,
		"group_id": gids,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate get total_size of asset error:%w", err)
	}
	err = DB.GetContext(ctx, &tsize, query, args...)
	if err != nil {
		return fmt.Errorf("get total_size of asset error:%w", err)
	}

	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query = fmt.Sprintf(`DELETE FROM %s WHERE user_id = ? AND id = ?`, tableNameAssetGroup)
	_, err = tx.ExecContext(ctx, query, userID, gid)
	if err != nil {
		return fmt.Errorf("delete asset group error:%w", err)
	}
	query = fmt.Sprintf(`UPDATE %s SET used_storage_size = used_storage_size - ? WHERE username=?`, tableNameUser)
	_, err = tx.ExecContext(ctx, query, tsize, userID)
	if err != nil {
		return fmt.Errorf("update user's used_storage_size error:%w", err)
	}

	return tx.Commit()
}

// DeleteOuterAssetGroup 删除指定的文件组，其内部的文件组先不进行删除
func DeleteOuterAssetGroup(ctx context.Context, userID string, gids []int64) error {
	query, args, err := squirrel.Delete(tableNameAssetGroup).Where(squirrel.Eq{
		"user_id": userID,
		"id":      gids,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of delete asset group error:%w", err)
	}
	_, err = DB.ExecContext(ctx, query, args)
	if err != nil {
		return fmt.Errorf("delete asset group error:%w", err)
	}

	return nil
}

// GetOnlyAssetsByUIDAndGroupID 获取用户文件组中唯一存在的文件区域映射
func GetOnlyAssetsByUIDAndGroupID(ctx context.Context, userID string, gids []int64) (map[string][]string, error) {
	var (
		as     []OnlyUserGroupAsset
		aaMaps = make(map[string][]string)
	)

	sb, sa, _ := squirrel.Select("`hash`").From(tableUserAsset).Where(squirrel.Eq{
		"user_id":  userID,
		"group_id": gids,
	}).ToSql()
	query, args, err := squirrel.Select("ua.cid,COUNT(ua.cid) AS num,uaa.area_id").From(fmt.Sprintf("%s AS uaa", tableUserAssetArea)).
		LeftJoin(fmt.Sprintf("%s AS ua ON uaa.`hash` = ua.`hash` AND ua.user_id = uaa.user_id", tableUserAsset)).
		Where(fmt.Sprintf("uaa.`hash` IN (%s)", sb), sa...).
		Where("ua.cid <> ''").GroupBy("ua.cid").Having("num = 1").ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get user_asset info error:%w", err)
	}
	err = DB.SelectContext(ctx, &as, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get user_asset info error:%w", err)
	}

	for _, v := range as {
		aaMaps[v.AreaID] = append(aaMaps[v.AreaID], v.CID)
	}

	return aaMaps, nil
}

// DeleteUserGroupAsset 删除用户文件组中的文件
func DeleteUserGroupAsset(ctx context.Context, userID string, gids []int64) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// 先通过user_asset的hash删除user_asset_area的数据
	sb, sa, _ := squirrel.Select("`hash`").From(tableUserAsset).Where(squirrel.Eq{
		"user_id":  userID,
		"group_id": gids,
	}).ToSql()
	query, args, err := squirrel.Delete(tableUserAssetArea).Where(fmt.Sprintf("`hash` IN (%s)", sb), sa...).
		Where("user_id = ?", userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of delete user_assest_area error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete user_assest_area error:%w", err)
	}
	// 删除user_asset的数据
	query, args, err = squirrel.Delete(tableUserAsset).Where(squirrel.Eq{
		"user_id":  userID,
		"group_id": gids,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of delete user_assest error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete user_assest error:%w", err)
	}
	// 删除当前文件组
	query, args, err = squirrel.Delete(tableNameAssetGroup).Where(squirrel.Eq{
		"user_id": userID,
		"id":      gids,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of delete user_assest_group error:%w", err)
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete user_assest_group error:%w", err)
	}

	return tx.Commit()
}

// GetUserGroupByParent 通过父级id获取其第一层的子级id
func GetUserGroupByParent(ctx context.Context, userID string, pids []int64) ([]int64, error) {
	var ids []int64

	query, args, err := squirrel.Select("id").From(tableNameAssetGroup).Where(squirrel.Eq{
		"user_id": userID,
		"parent":  pids,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get user_assest group error:%w", err)
	}
	err = DB.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get user_assest group error:%w", err)
	}

	return ids, nil
}

// AddUserAssetMap 增加用户文件映射表
func AddUserAssetMap(ctx context.Context, userID, hash string) error {
	query, args, err := squirrel.Insert(tableUserAssetMap).Columns("user_id", "asset_hash").Values(userID, hash).Options("IGNORE").ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of insert user_asset_map error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert user_asset_map error:%w", err)
	}

	return nil
}

// AddAssetGroupVisitCount 增加文件组分享次数
func AddAssetGroupVisitCount(ctx context.Context, userID string, id int) error {
	query, args, err := squirrel.Update(tableNameAssetGroup).Set("visit_count", squirrel.Expr("visit_count + 1")).Where("id = ? AND user_id = ?", id, userID).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of update asset group's visit count error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update asset group's visit count error:%w", err)
	}

	return nil
}

// CheckAssetByMd5AndAreaExists 判断文件是否已经存在
func CheckAssetByMd5AndAreaExists(ctx context.Context, md5, areaID string) (bool, error) {
	var hash string

	if strings.TrimSpace(md5) == "" {
		return false, nil
	}

	// 通过md5或者hash
	query, args, err := squirrel.Select("hash").From(tableUserAsset).Where("md5 = ?", md5).Limit(1).ToSql()
	if err != nil {
		return false, fmt.Errorf("generate select hash from user_asset error:%w", err)
	}
	err = DB.GetContext(ctx, &hash, query, args...)
	if err != nil {
		return false, fmt.Errorf("get hash from user_asset error:%w", err)
	}

	query, args, err = squirrel.Select("hash").From(tableUserAssetArea).Where("hash = ? AND area_id = ? AND is_sync = 1", hash, areaID).Limit(1).ToSql()
	if err != nil {
		return false, fmt.Errorf("generate select hash from user_asset_area error:%w", err)
	}
	err = DB.GetContext(ctx, &hash, query, args...)
	if err != nil {
		return false, fmt.Errorf("get hash from user_asset_area error:%w", err)
	}

	return true, nil
}

// GetNoExistCIDs 获取用户不存在的cid信息
func GetNoExistCIDs(ctx context.Context, uid string, cids []string) ([]string, error) {
	var ncids []string

	query, args, err := squirrel.Select("cid").From(tableUserAsset).Where(squirrel.NotEq{"cid": cids}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get not exists cid error:%w", err)
	}

	err = DB.SelectContext(ctx, &ncids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get not exists cid error:%w", err)
	}

	return ncids, nil
}
