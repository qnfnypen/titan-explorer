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

	_, err = tx.ExecContext(ctx, updateRewardQuery, statement.Amount, statement.UserName)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (username, from, to, amount, event, status, created_at, updated_at)
			VALUES (:username, :from, :to, :amount, :event, :status, :created_at, :updated_at);`, tableNameRewardStatement),
		statement)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetRewardStatementByFrom(ctx context.Context, from string) (*model.RewardStatement, error) {
	var out model.RewardStatement

	query := fmt.Sprintf("select * from %s where from = ?", tableNameRewardStatement)
	err := DB.QueryRowxContext(ctx, query, from).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRow
		}
		return nil, err
	}

	return &out, nil
}
