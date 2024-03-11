package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/jmoiron/sqlx"
	"time"
)

const (
	tableNameRewardStatement = "reward_statement"
	tableNameRewardWithdraw  = "withdraw_record"
)

func UpdateUserRewardOld(ctx context.Context, statement *model.RewardStatement) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if statement.Event == model.RewardEventEarning || statement.Event == model.RewardEventReferrals {
		err = insertOrUpdateRewardStatement(ctx, tx, statement)
		if err != nil {
			return err
		}
		return tx.Commit()
	}

	updateRewardQuery := fmt.Sprintf("update %s set reward = reward + ? where username = ?", tableNameUser)

	_, err = tx.ExecContext(ctx, updateRewardQuery, statement.Amount, statement.FromUser)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (username, from_user, amount, event, status, device_id, created_at, updated_at)
			VALUES (:username, :from_user, :amount, :event, :status, :device_id, :created_at, :updated_at);`, tableNameRewardStatement)

	_, err = tx.NamedExecContext(ctx, query, statement)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func UpdateUserReward(ctx context.Context, user *model.User) error {
	updateRewardQuery := fmt.Sprintf("update %s set reward = ?, referral_reward = ?, device_count = ? where username = ?", tableNameUser)
	_, err := DB.ExecContext(ctx, updateRewardQuery, user.Reward, user.RefereralReward, user.DeviceCount, user.Username)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUserReferralReward(ctx context.Context, user *model.User) error {
	updateRewardQuery := fmt.Sprintf("update %s set referral_reward = ? where username = ?", tableNameUser)
	_, err := DB.ExecContext(ctx, updateRewardQuery, user.RefereralReward, user.Username)
	if err != nil {
		return err
	}
	return nil
}

func insertOrUpdateRewardStatement(ctx context.Context, tx *sqlx.Tx, statement *model.RewardStatement) error {
	getQuery := fmt.Sprintf(`select * from %s where username = ? and event =? and created_at >= ? limit 1`, tableNameRewardStatement)

	var rs model.RewardStatement
	err := DB.GetContext(ctx, &rs, getQuery, statement.Username, statement.Event, statement.CreatedAt.Format(time.DateOnly))
	if errors.Is(err, sql.ErrNoRows) {
		updateRewardQuery := fmt.Sprintf("update %s set reward = reward + ? where username = ?", tableNameUser)

		_, err = tx.ExecContext(ctx, updateRewardQuery, statement.Amount, statement.FromUser)
		if err != nil {
			return err
		}

		query := fmt.Sprintf(
			`INSERT INTO %s (username, from_user, amount, event, status, device_id, created_at, updated_at)
			VALUES (:username, :from_user, :amount, :event, :status, :device_id, :created_at, :updated_at);`, tableNameRewardStatement)

		_, err = tx.NamedExecContext(ctx, query, statement)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	reward := statement.Amount - rs.Amount
	updateRewardQuery := fmt.Sprintf("update %s set reward = reward + ? where username = ?", tableNameUser)
	_, err = tx.ExecContext(ctx, updateRewardQuery, reward, statement.FromUser)
	if err != nil {
		return err
	}

	updateQuery := fmt.Sprintf("update %s set amount = ? where id = ?", tableNameRewardStatement)
	_, err = tx.ExecContext(ctx, updateQuery, statement.Amount, rs.ID)
	if err != nil {
		return err
	}

	return nil
}

func GetRewardStatementByDeviceID(ctx context.Context, deviceId string) (*model.RewardStatement, error) {
	var out model.RewardStatement

	query := fmt.Sprintf("select * from %s where device_id = ?", tableNameRewardStatement)
	err := DB.QueryRowxContext(ctx, query, deviceId).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRow
		}
		return nil, err
	}

	return &out, nil
}

func GetReferralList(ctx context.Context, username string, option QueryOption) (int64, []*model.InviteFrensRecord, error) {
	var out []*model.InviteFrensRecord

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64

	countQuery := `select count(1) from users where referrer in (select referral_code from users where username = ?);`
	countQueryIn, countQueryParams, err := sqlx.In(countQuery, username)
	if err != nil {
		return 0, nil, err
	}

	err = DB.GetContext(ctx, &total, countQueryIn, countQueryParams...)
	if err != nil {
		return 0, nil, err
	}

	query := `select username as email, device_count as bound_count, (reward * 0.05) as reward, created_at as time from users where referrer in (select referral_code from users where username = ?) order by created_at desc LIMIT ? OFFSET ?;`

	err = DB.SelectContext(ctx, &out, query, username, limit, offset)
	if err != nil {
		return 0, nil, err
	}

	return total, out, nil

}

func GetReferralListOld(ctx context.Context, username string, events []model.RewardEvent, option QueryOption) (int64, []*model.InviteFrensRecord, error) {
	var out []*model.InviteFrensRecord

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64

	countQuery := fmt.Sprintf(`SELECT count(distinct from_user) FROM %s where event in (?) and username = ?`, tableNameRewardStatement)
	countQueryIn, countQueryParams, err := sqlx.In(countQuery, events, username)
	if err != nil {
		return 0, nil, err
	}

	err = DB.GetContext(ctx, &total, countQueryIn, countQueryParams...)
	if err != nil {
		return 0, nil, err
	}

	subQuery := fmt.Sprintf(`select from_user as email, count(distinct rs.event) as status, SUM(IF(rs.event = 'bind_device', 1, 0)) as bound_count, sum(rs.amount) as reward, min(created_at) as time 
			from %s rs where event in (?) and username = ? group by from_user`, tableNameRewardStatement)

	subQueryIn, subQueryParams, err := sqlx.In(subQuery, events, username)
	if err != nil {
		return 0, nil, err
	}

	query := fmt.Sprintf(`select * from (%s) s order by time DESC LIMIT ? OFFSET ?`, subQueryIn)

	err = DB.SelectContext(ctx, &out, query, append(subQueryParams, limit, offset)...)
	if err != nil {
		return 0, nil, err
	}

	return total, out, nil
}

//func GetUserReferralReward(ctx context.Context, username string) (int64, error) {
//	var total int64
//	query := fmt.Sprintf(`select ifnull(sum(amount),0) from %s where event in ("invite_frens", "bind_device") and username = ?`, tableNameRewardStatement)
//
//	err := DB.GetContext(ctx, &total, query, username)
//	if err != nil {
//		return 0, err
//	}
//
//	return total, nil
//}

func AddWithdrawRequest(ctx context.Context, withdraw *model.Withdraw) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateRewardQuery := fmt.Sprintf("update %s set reward = reward - ?, frozen_reward = frozen_reward + ? where username = ?", tableNameUser)

	_, err = tx.ExecContext(ctx, updateRewardQuery, withdraw.Amount, withdraw.Amount, withdraw.Username)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (username, to_address, amount, hash, status, created_at, updated_at)
			VALUES (:username, :to_address, :amount, :hash, :status, :created_at, :updated_at);`, tableNameRewardWithdraw),
		withdraw)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetWithdrawRecordList(ctx context.Context, username string, option QueryOption) (int64, []*model.Withdraw, error) {
	var out []*model.Withdraw

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64

	countQuery := fmt.Sprintf(`SELECT count(*) FROM %s where username = ?`, tableNameRewardWithdraw)

	err := DB.GetContext(ctx, &total, countQuery, username)
	if err != nil {
		return 0, nil, err
	}

	query := fmt.Sprintf(`select * from %s where username = ? order by created_at DESC LIMIT ? OFFSET ?`, tableNameRewardWithdraw)

	err = DB.SelectContext(ctx, &out, query, username, limit, offset)
	if err != nil {
		return 0, nil, err
	}

	return total, out, nil
}
