package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func AddKOL(ctx context.Context, kol *model.KOL) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `INSERT INTO kol(user_id, level, comment, status, created_at, updated_at) VALUES (:user_id, :level, :comment, :status, now(), now());`, kol)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `update users set role = ?, updated_at = now() where username = ?`, model.UserRoleKOL, kol.UserId)
	if err != nil {
		return err
	}

	return tx.Commit()
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
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userRole := model.UserRoleDefault
	if kol.Status == 1 {
		userRole = model.UserRoleKOL
	}

	_, err = DB.ExecContext(ctx,
		`update kol set level = ?, comment = ?, status = ?, updated_at = now() where user_id = ?`,
		kol.Level, kol.Comment, kol.Status, kol.UserId,
	)

	_, err = tx.ExecContext(ctx, `update users set role = ?, updated_at = now() where username = ?`, userRole, kol.UserId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func UpdateKOLLevel(ctx context.Context, kolUserId string, level int) error {
	_, err := DB.ExecContext(ctx,
		`update kol set level = ?, updated_at = now() where user_id = ?`, level, kolUserId,
	)
	return err
}

func GetAllKOLLevels(ctx context.Context) (map[string]*model.KOLLevel, error) {
	query := `select t1.user_id, t1.level, t2.parent_commission_percent, t2.children_bonus_percent, t2.device_threshold from kol t1 left join kol_level_conf t2 on t1.level = t2.level where t1.status = 1 and t2.status = 1`
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

func GetKolList(ctx context.Context, option QueryOption) ([]*model.KOL, int64, error) {
	var total int64
	var out []*model.KOL

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	err := DB.GetContext(ctx, &total, `SELECT count(*) FROM kol`)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM kol LIMIT %d OFFSET %d`, limit, offset,
	))
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func AddKOLLevelConfig(ctx context.Context, levelConf *model.KOLLevelConf) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO kol_level_conf(level, parent_commission_percent, children_bonus_percent, user_threshold, device_threshold, status, created_at, updated_at) 
		VALUES (:level, :parent_commission_percent, :children_bonus_percent, :user_threshold, :device_threshold, :status,  now(), now());`, levelConf)
	return err
}

func GetKOLLevelByLevel(ctx context.Context, level int) (*model.KOLLevelConf, error) {
	var out model.KOLLevelConf
	err := DB.GetContext(ctx, &out, `select * from kol_level_conf where level = ?`, level)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdateKOLLevelConfig(ctx context.Context, levelConf *model.KOLLevelConf) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userRole := model.UserRoleDefault
	if levelConf.Status == 1 {
		userRole = model.UserRoleKOL
	}

	_, err = tx.ExecContext(ctx,
		`update kol_level_conf set parent_commission_percent = ?, children_bonus_percent = ?, user_threshold = ?, device_threshold = ?,  status = ?,  updated_at = now() where level = ?`,
		levelConf.ParentCommissionPercent, levelConf.ChildrenBonusPercent, levelConf.UserThreshold, levelConf.DeviceThreshold, levelConf.Status, levelConf.Level,
	)

	_, err = tx.ExecContext(ctx, `update users set role = ?, updated_at = now() where username in (select user_id from kol where level = ?)`, userRole, levelConf.Level)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func DeleteKOLLevelConfig(ctx context.Context, level int) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`delete from kol_level_conf where level = ?`, level,
	)

	_, err = tx.ExecContext(ctx, `update users set role = ?, updated_at = now() where username in (select user_id from kol where level = ?)`, model.UserRoleDefault, level)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func DeleteKOL(ctx context.Context, userId string) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`delete from kol where user_id = ?`, userId,
	)

	_, err = tx.ExecContext(ctx, `update users set role = ?, updated_at = now() where username = ?`, model.UserRoleDefault, userId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetKolLevelConfig(ctx context.Context, option QueryOption) ([]*model.KOLLevelConf, int64, error) {
	var total int64
	var out []*model.KOLLevelConf

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	err := DB.GetContext(ctx, &total, `SELECT count(*) FROM kol_level_conf`)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM kol_level_conf order by level LIMIT %d OFFSET %d`, limit, offset,
	))
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func AddKOLLevelUPRecord(ctx context.Context, record *model.KOLLevelUPRecord) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO kol_level_up_record( user_id, before_level,after_level, referral_users_count,referral_nodes_count, created_at) 
		VALUES (:user_id, :before_level, :after_level, :referral_users_count, :referral_nodes_count, :created_at);`, record)
	return err
}
