package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameRewardStatement = "reward_statement"

func UpdateUserReward(ctx context.Context, statement *model.RewardStatement) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateRewardQuery := fmt.Sprintf("update %s set reward = reward + ? where username = ?", tableNameUser)

	_, err = tx.ExecContext(ctx, updateRewardQuery, statement.Amount, statement.Recipient)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (recipient, username, amount, event, status, device_id, created_at, updated_at)
			VALUES (:recipient, ;username, :amount, :event, :status, :device_id, :created_at, :updated_at);`, tableNameRewardStatement),
		statement)
	if err != nil {
		return err
	}

	return tx.Commit()
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

func GetReferralList(ctx context.Context, recipient string, option QueryOption) (int64, []*model.InviteFrensRecord, error) {
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

	countQuery := fmt.Sprintf(`SELECT count(distinct username) FROM %s where event in ("invite_frens", "bind_device") and recipient = ?`, tableNameRewardStatement)

	err := DB.GetContext(ctx, &total, countQuery, recipient)
	if err != nil {
		return 0, nil, err
	}

	subQuery := fmt.Sprintf(`select username as email, count(distinct rs.event) as status, SUM(IF(rs.event = 'bind_device', 1, 0)) as bound_count, sum(rs.amount) as reward, min(created_at) as time 
			from %s rs where event in ("invite_frens", "bind_device") and recipient = ? group by username`, tableNameRewardStatement)

	query := fmt.Sprintf(`select * from (%s) s order by time DESC LIMIT ? OFFSET ?`, subQuery)

	err = DB.SelectContext(ctx, &out, query, recipient, limit, offset)
	if err != nil {
		return 0, nil, err
	}

	return total, out, nil
}
