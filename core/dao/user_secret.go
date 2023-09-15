package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameUserSecret = "user_secret"

func AddUserSecret(ctx context.Context, userSecret *model.UserSecret) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (user_id, app_key, app_secret, status, created_at, updated_at) 
			VALUES (:user_id, :app_key, :app_secret, :status, :created_at, :updated_at);`, tableNameUserSecret),
		userSecret)
	return err
}

func GetSecretKey(ctx context.Context, key string) (*model.UserSecret, error) {
	var secret model.UserSecret
	err := DB.GetContext(ctx, &secret, fmt.Sprintf(
		`SELECT * from %s WHERE app_secret = ?`, tableNameUserSecret), key)
	if err != nil {
		return nil, err
	}
	return &secret, err
}
