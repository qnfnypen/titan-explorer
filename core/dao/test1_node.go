package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	test1NodeTable = "device_info"
	zeroTime       = "0000-00-00 00:00:00.000"
)

type UserAndQuest struct {
	ID                     int64     `db:"id" json:"id"`
	Uuid                   string    `db:"uuid" json:"uuid"`
	Avatar                 string    `db:"avatar" json:"avatar"`
	Username               string    `db:"username" json:"username"`
	PassHash               string    `db:"pass_hash" json:"-"`
	UserEmail              string    `db:"user_email" json:"user_email"`
	WalletAddress          string    `db:"wallet_address" json:"wallet_address"`
	Role                   int32     `db:"role" json:"role"`
	AllocateStorage        int       `db:"allocate_storage" json:"allocate_storage"`
	ProjectId              int64     `db:"project_id"`
	Referrer               string    `db:"referrer" json:"referrer"`
	ReferrerUserId         string    `db:"referrer_user_id" json:"-"`
	ReferralCode           string    `db:"referral_code" json:"referral_code"`
	Reward                 float64   `db:"reward" json:"reward"`
	ReferralReward         float64   `db:"referral_reward" json:"referral_reward"`
	ClosedTestReward       float64   `db:"closed_test_reward" json:"closed_test_reward"`
	HuygensReward          float64   `db:"huygens_reward" json:"huygens_reward"`
	HuygensReferralReward  float64   `db:"huygens_referral_reward" json:"huygens_referral_reward"`
	HerschelReward         float64   `db:"herschel_reward" json:"herschel_reward"`
	HerschelReferralReward float64   `db:"herschel_referral_reward" json:"herschel_referral_reward"`
	CassiniReward          float64   `db:"cassini_reward" json:"cassini_reward"`
	CassiniReferralReward  float64   `db:"cassini_referral_reward" json:"cassini_referral_reward"`
	DeviceCount            int64     `db:"device_count" json:"device_count"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	HerschelCredits        int64     `json:"herschel_credits" db:"-"`
	HerschelInviteCredits  int64     `json:"herschel_invite_credits" db:"-"`
	CassiniCredits         int64     `json:"cassini_credits" db:"-"`
	CassiniInviteCredits   int64     `json:"cassini_invite_credits" db:"-"`
	OnlineIncentiveReward  float64   `json:"online_incentive_reward" db:"-"`
	UpdatedAt              time.Time `db:"updated_at" json:"-"`
	DeletedAt              time.Time `db:"deleted_at" json:"-"`
}

type (
	// NodeStatusInfo 节点状态信息
	NodeStatusInfo struct {
		Name     string `db:"device_name" json:"name"`
		AreaID   string `db:"area_id" json:"area_id"`
		DeviceID string `db:"device_id" json:"node_id"`
		Status   int64  `db:"status" json:"status"` // 1-在线 2-故障 3-离线 11-已退出
		ExpTime  int64  `db:"deactive_time" json:"exp_time"`
	}
	// NodeStatus 节点状态数量
	NodeStatus struct {
		Status int64 `db:"status"` // 1-在线 2-故障 3-离线 11-已退出
		Num    int64 `db:"num"`
	}
)

// GetTest1Nodes 获取test1节点信息
func GetTest1Nodes(ctx context.Context, statusCode int64, page, size uint64) (int64, []model.Test1NodeInfo, error) {
	// device_status_code 1-在线 2-故障 3-离线
	var (
		totalBuilder squirrel.SelectBuilder
		infoBuilder  squirrel.SelectBuilder

		total int64
		infos = make([]model.Test1NodeInfo, 0)
	)

	if statusCode <= 0 || statusCode > 4 {
		return 0, nil, errors.New("param error")
	}

	// 获取删除节点
	if statusCode == 4 {
		totalBuilder = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("deleted_at <> ?", 0)
		infoBuilder = squirrel.Select("device_name,external_ip,system_version,device_id,ip_location,cumulative_profit").From(test1NodeTable).
			Where("deleted_at <> ?", 0).Offset((page - 1) * size).Limit(size)
	} else {
		totalBuilder = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = ?", statusCode)
		infoBuilder = squirrel.Select("device_name,external_ip,system_version,device_id,ip_location,cumulative_profit").From(test1NodeTable).
			Where("device_status_code = ?", statusCode).Offset((page - 1) * size).Limit(size)
	}

	query, args, err := totalBuilder.ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get total of device's info error:%w", err)
	}

	query, args, err = infoBuilder.ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get device's info error:%w", err)
	}
	err = DB.SelectContext(ctx, &infos, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf(" get device's info error:%w", err)
	}

	return total, infos, nil
}

// GetNodeNums 获取节点总数
func GetNodeNums(ctx context.Context, usedID string) (int64, int64, int64, int64, error) {
	var online, abnormal, offline, deleted int64
	// device_status_code 1-在线 2-故障 3-离线
	// 在线
	query, args, err := squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 1 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &online, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 故障
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 2 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &abnormal, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 离线
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 3 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &offline, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 删除
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("deleted_at <> 0 AND user_id = ?", usedID).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &deleted, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}

	return online, abnormal, offline, deleted, nil
}

// UpdateTest1DeviceName 编辑节点设备备注
func UpdateTest1DeviceName(ctx context.Context, id, name string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("device_name", name).Where("device_id = ?", id).ToSql()
	if err != nil {
		return fmt.Errorf("generate update device's name error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("update device's name error:%w", err)
	}

	return nil
}

// DeleteOfflineDevice 删除离线设备
func DeleteOfflineDevice(ctx context.Context, ids []string, usedID string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("deleted_at", time.Now()).Where(squirrel.Eq{
		"device_id":          ids,
		"user_id":            usedID,
		"device_status_code": 3,
		"deleted_at":         zeroTime,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate delete offline's device error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("delete offline's device error:%w", err)
	}

	return nil
}

// MoveBackDeletedDevice 移回删除的设备
func MoveBackDeletedDevice(ctx context.Context, ids []string, usedID string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("deleted_at", zeroTime).Where("deleted_at <> 0").Where(squirrel.Eq{
		"device_id": ids,
		"user_id":   usedID,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate move back deleted device error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("move back deleted device error:%w", err)
	}

	return nil
}

// GetCreditByUn 获取社区奖励
func GetCreditByUn(ctx context.Context, un string) (int64, error) {
	var credit int64

	query, args, err := squirrel.Select("IFNULL(SUM(credit),0)").From("user_mission").Where("username = ?", un).ToSql()
	if err != nil {
		return 0, fmt.Errorf("get sum of credit error:%w", err)
	}

	err = QDB.GetContext(ctx, &credit, query, args...)
	if err != nil {
		return 0, fmt.Errorf("get credit of user's mission error:%w", err)
	}

	return credit, nil
}

// GetInviteCreditByUn 获取社区邀请奖励
func GetInviteCreditByUn(ctx context.Context, un string) (int64, error) {
	var credit int64

	query, args, err := squirrel.Select("IFNULL(SUM(credit),0)").From("invite_log").Where("username = ?", un).ToSql()
	if err != nil {
		return 0, fmt.Errorf("get sum of credit error:%w", err)
	}

	err = QDB.GetContext(ctx, &credit, query, args...)
	if err != nil {
		return 0, fmt.Errorf("get credit of user's mission error:%w", err)
	}

	return credit, nil
}

// GetOnlineNodes 获取在线的节点数量
func GetOnlineNodes(ctx context.Context) (int64, error) {
	var online int64

	query, args, err := squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 1 AND deleted_at = ?", zeroTime).ToSql()
	if err != nil {
		return 0, fmt.Errorf("generate sql error:%w", err)
	}
	err = DB.GetContext(ctx, &online, query, args...)
	if err != nil {
		return 0, fmt.Errorf("get online node error:%w", err)
	}

	return online, nil
}

// GetNodeInfos 获取用户节点信息
func GetNodeInfos(ctx context.Context, uid string, page, size uint64) ([]NodeStatus, []NodeStatusInfo, error) {
	var (
		statusNums  []NodeStatus
		statusInfos []NodeStatusInfo
	)

	query, args, err := squirrel.Select("IF(operation>0,operation+10,device_status_code) AS `status`,COUNT(device_id) AS num").
		From(test1NodeTable).Where("user_id = ? AND node_type = 2", uid).GroupBy("`status`").ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("generate sql of get device status number error:%w", err)
	}
	err = DB.SelectContext(ctx, &statusNums, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("get device status number error:%w", err)
	}

	query, args, err = squirrel.Select("IF(operation>0,operation+10,device_status_code) AS `status`,device_name,area_id,device_id,deactive_time").
		From(test1NodeTable).Where("user_id = ? AND node_type = 2", uid).Offset((page - 1) * size).Limit(size).ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("generate sql of get device status info error:%w", err)
	}
	err = DB.SelectContext(ctx, &statusInfos, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("get device status info error:%w", err)
	}

	return statusNums, statusInfos, nil
}

// CheckIsNodeOwner 校验是否为节点拥有者
func CheckIsNodeOwner(ctx context.Context, uid, nodeID string) (int64, error) {
	var status int64

	query, args, err := squirrel.Select("IF(operation>0,operation+10,device_status_code) AS `status`").From(test1NodeTable).Where("user_id = ? AND device_id = ?", uid, nodeID).ToSql()
	if err != nil {
		return 0, fmt.Errorf("generate sql of get user_id of device_info error:%w", err)
	}

	err = DB.GetContext(ctx, &status, query, args...)
	switch err {
	case nil:
		return status, nil
	case sql.ErrNoRows:
		return 0, nil
	default:
		return 0, fmt.Errorf("get user_id of device_info error:%w", err)
	}
}

// UpdateNodeOperationStatus 修改节点操作状态
func UpdateNodeOperationStatus(ctx context.Context, uid, nodeID string, operation int64, hours ...int) error {
	var hour int

	if len(hours) > 0 {
		hour = hours[0]
	}
	maps := make(map[string]interface{})
	maps["operation"] = operation
	if operation != 0 {
		maps["deactive_time"] = time.Now().Add(time.Duration(hour) * time.Hour).Unix()
	}

	query, args, err := squirrel.Update(test1NodeTable).Where("user_id = ? AND device_id = ?", uid, nodeID).SetMap(maps).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of update operation of device_info error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update operation of device_info error:%w", err)
	}

	return nil
}
