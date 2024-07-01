package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func AddKOL(ctx context.Context, kol *model.KOL) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO kol(user_id, level, comment, status, created_at, updated_at) VALUES (:user_id, :level, :comment, :status, now(), now());`, kol)
	if err != nil {
		return err
	}

	return err
}

func UpsertKOLs(ctx context.Context, kols []*model.KOL) error {
	query := `INSERT INTO kol(user_id, level, comment, status, created_at, updated_at) VALUES (:user_id, :level, :comment, :status, now(), now()) 
		ON DUPLICATE KEY UPDATE level = values(level), updated_at = now();`

	_, err := DB.NamedExecContext(ctx, query, kols)
	if err != nil {
		return err
	}

	return err
}

func GetKOLByUserId(ctx context.Context, userId string) (*model.KOL, error) {
	var out model.KOL
	err := DB.GetContext(ctx, &out, `select * from kol where user_id = ?`, userId)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdateKOL(ctx context.Context, kol *model.KOL) error {
	_, err := DB.ExecContext(ctx,
		`update kol set level = ?, comment = ?, status = ?, updated_at = now() where user_id = ?`,
		kol.Level, kol.Comment, kol.Status, kol.UserId,
	)
	return err
}

func UpdateKOLLevel(ctx context.Context, kolUserId string, level int) error {
	_, err := DB.ExecContext(ctx,
		`update kol set level = ?, updated_at = now() where user_id = ?`, level, kolUserId,
	)
	return err
}

func GetAllKOLLevels(ctx context.Context) (map[string]*model.KOLLevel, error) {
	query := `select t1.user_id, t1.level, t2.parent_commission_percent, t2.children_bonus_percent, t2.device_threshold from kol t1 left join kol_level_config t2 on t1.level = t2.level where t1.status = 1 and t2.status = 1`
	rows, err := DB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]*model.KOLLevel)

	for rows.Next() {
		var kl model.KOLLevel
		err = rows.StructScan(&kl)
		if err != nil {
			log.Errorf("scan %v", err)
			continue
		}
		out[kl.UserId] = &kl
	}

	return out, nil
}

func GetAdminAddedKolLevel(ctx context.Context) (map[string]int, error) {
	query := `select user_id, level from kol where comment <> 'system'`

	out := make(map[string]int)
	rows, err := DB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userId string
		var level int
		err = rows.Scan(&userId, &level)
		if err != nil {
			return nil, err
		}

		out[userId] = level
	}

	return out, nil
}

func GetKolList(ctx context.Context, option QueryOption) ([]*model.KOL, int64, error) {
	var total int64
	var out []*model.KOL

	var (
		args  []interface{}
		where = " where 1=1"
	)

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	if option.UserID != "" {
		where += " and user_id = ?"
		args = append(args, option.UserID)
	}

	err := DB.GetContext(ctx, &total, `SELECT count(1) FROM kol`+where, args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM kol %s LIMIT %d OFFSET %d`, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func AddKOLLevelConfig(ctx context.Context, levelConf *model.KOLLevelConfig) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO kol_level_config(level, parent_commission_percent, children_bonus_percent, user_threshold, device_threshold, status, created_at, updated_at) 
		VALUES (:level, :parent_commission_percent, :children_bonus_percent, :user_threshold, :device_threshold, :status,  now(), now());`, levelConf)
	return err
}

func GetKOLLevelByLevel(ctx context.Context, level int) (*model.KOLLevelConfig, error) {
	var out model.KOLLevelConfig
	err := DB.GetContext(ctx, &out, `select * from kol_level_config where level = ?`, level)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdateKOLLevelConfig(ctx context.Context, levelConf *model.KOLLevelConfig) error {
	_, err := DB.ExecContext(ctx,
		`update kol_level_config set commission_percent = ?, parent_commission_percent = ?, device_threshold = ?,  status = ?,  updated_at = now() where level = ?`,
		levelConf.CommissionPercent, levelConf.ParentCommissionPercent, levelConf.DeviceThreshold, levelConf.Status, levelConf.Level,
	)

	return err
}

func DeleteKOLLevelConfig(ctx context.Context, level int) error {
	_, err := DB.ExecContext(ctx,
		`delete from kol_level_config where level = ?`, level,
	)

	return err
}

func DeleteKOL(ctx context.Context, userId string) error {
	_, err := DB.ExecContext(ctx,
		`delete from kol where user_id = ?`, userId,
	)

	return err
}

func GetKolLevelConfig(ctx context.Context, option QueryOption) ([]*model.KOLLevelConfig, int64, error) {
	var total int64
	var out []*model.KOLLevelConfig

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	err := DB.GetContext(ctx, &total, `SELECT count(*) FROM kol_level_config`)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM kol_level_config order by level LIMIT %d OFFSET %d`, limit, offset,
	))
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
