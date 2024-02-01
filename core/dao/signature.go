package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameSignature = "signature"

func AddSignature(ctx context.Context, signature *model.Signature) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (username, node_id, message, hash, signature, created_at, updated_at) VALUES (:username, :node_id, :message, :hash, :signature, now(), now());`, tableNameSignature),
		signature)
	return err
}

func GetSignatureByHash(ctx context.Context, hash string) (*model.Signature, error) {
	var out model.Signature
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE hash = ?`, tableNameSignature), hash,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return &out, nil
}
