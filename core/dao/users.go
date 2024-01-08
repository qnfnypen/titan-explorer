package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameUser = "users"

func CreateUser(ctx context.Context, user *model.User) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (uuid, username, pass_hash, user_email, wallet_address, role, referrer, referral_code, created_at)
			VALUES (:uuid, :username, :pass_hash, :user_email, :wallet_address, :role, :referrer, :referral_code, :created_at);`, tableNameUser,
	), user)
	return err
}

func ResetPassword(ctx context.Context, passHash, username string) error {
	_, err := DB.DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET pass_hash = '%s', updated_at = now() WHERE username = '%s'`, tableNameUser, passHash, username))
	return err
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var out model.User
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE username = ?`, tableNameUser), username,
	).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func GetUserIds(ctx context.Context) ([]string, error) {
	queryStatement := fmt.Sprintf(`SELECT username as user_id FROM %s;`, tableNameUser)
	var out []string
	err := DB.SelectContext(ctx, &out, queryStatement)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return out, nil
}

func UpdateUserWalletAddress(ctx context.Context, username, address string) error {
	query := fmt.Sprintf("update %s set wallet_address = ? where username = ?", tableNameUser)
	_, err := DB.ExecContext(ctx, query, address, username)
	return err
}

func GetUsersReferrer(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	query := fmt.Sprintf("SELECT * FROM %s u JOIN %s ur on ur.referral_code=ur.referrer WHERE ur.username=? LIMIT 1", tableNameUser, tableNameUser)
	err := DB.QueryRowxContext(ctx, query, username).StructScan(&u)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return &u, nil
}

func GetUserByRefCode(ctx context.Context, refCode string) (*model.User, error) {
	var u model.User
	query := "SELECT * FROM users WHERE referral_code=? LIMIT 1"
	err := DB.QueryRowxContext(ctx, query, refCode).Scan(&u)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return &u, nil
}
