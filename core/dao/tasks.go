package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

var tableNameTaskInfo = "task_info"

func GetTaskInfoByCID(ctx context.Context, cid string) (*model.TaskInfo, error) {
	var out model.TaskInfo
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE cid = ?`, tableNameTaskInfo), cid,
	).StructScan(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func GetTaskInfoByTime(ctx context.Context, deviceID, cid string, time time.Time) (*model.TaskInfo, error) {
	var out model.TaskInfo
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id =? AND cid = ? AND time = ?`, tableNameTaskInfo),
		deviceID, cid, time,
	).StructScan(&out); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &out, nil
}

func CreateTaskInfo(ctx context.Context, task *model.TaskInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (user_id, miner_id, device_id, file_name, ip_address, cid, bandwidth_up,
			bandwidth_down, time_need, time, service_country, region, status, price, file_size, download_url,
			created_at, updated_at, deleted_at)
			VALUES (:user_id, :miner_id, :device_id, :file_name, :ip_address, :cid, :bandwidth_up, :bandwidth_down, :time_need,
			    :time, :service_country, :region, :status, :price, :file_size, :download_url, :created_at, :updated_at, :deleted_at);`, tableNameTaskInfo,
	), task)
	return err
}

func UpsertTaskInfo(ctx context.Context, task *model.TaskInfo) error {
	old, err := GetTaskInfoByCID(ctx, task.Cid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if old == nil {
		return CreateTaskInfo(ctx, task)
	}

	old.BandwidthUp = task.BandwidthUp
	old.BandwidthDown = task.BandwidthDown
	old.TimeNeed = task.TimeNeed
	old.Price = task.Price
	return UpdateTaskInfo(ctx, old)
}

func UpdateTaskInfo(ctx context.Context, task *model.TaskInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET "bandwidth_up" =:bandwidth_up, "bandwidth_down" =:bandwidth_down, "time_need" =:time_need, 
              "price" =:price, "updated_at" =:updated_at WHERE "id" = :id`, tableNameTaskInfo),
		task)
	return err
}

func GetTaskInfoList(ctx context.Context, cond *model.TaskInfo, option QueryOption) ([]*model.TaskInfo, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.Status != "" && cond.Status != "All" {
		where += ` AND status = ?`
		args = append(args, cond.Status)
	}
	if cond.Cid != "" {
		where += ` AND cid = ?`
		args = append(args, cond.Cid)
	}
	if cond.UserID != "" {
		where += ` AND user_id = ?`
		args = append(args, cond.UserID)
	}

	limit := option.PageSize
	offset := option.Page
	if option.Page > 0 {
		offset = option.PageSize * (option.Page - 1)
	}
	if option.PageSize <= 0 {
		limit = 50
	}

	var total int64
	var out []*model.TaskInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameTaskInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameTaskInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
