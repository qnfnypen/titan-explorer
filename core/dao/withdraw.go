package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/jmoiron/sqlx"
)

const (
	tableNameRewardWithdraw = "withdraw_record"
)

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

	countQuery := `select count(1) from users where referrer_user_id = ?;`
	countQueryIn, countQueryParams, err := sqlx.In(countQuery, username)
	if err != nil {
		return 0, nil, err
	}

	err = DB.GetContext(ctx, &total, countQueryIn, countQueryParams...)
	if err != nil {
		return 0, nil, err
	}

	query := `select username as email, device_count as bound_count, if(device_count>0, 2, 1) as status, referrer_commission_reward as reward, referrer, created_at as time from users where referrer_user_id = ? order by created_at desc LIMIT ? OFFSET ?;`

	err = DB.SelectContext(ctx, &out, query, username, limit, offset)
	if err != nil {
		return 0, nil, err
	}

	return total, out, nil

}

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
