package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameUser = "users"

func CreateUser(ctx context.Context, user *model.User) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (uuid, username, pass_hash, user_email, address, role)
			VALUES (:uuid, :username, :pass_hash, :user_email, :address, :role);`, tableNameUser,
	), user)
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

func GetUserByUserUUID(ctx context.Context, UUID string) (*model.User, error) {
	var out model.User
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE uuid = ?`, tableNameUser), UUID,
	).StructScan(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
