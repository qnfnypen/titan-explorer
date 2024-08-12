package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var (
	tableEdgeConfig = "edge_config"
)

func GetEdgeConfig(ctx context.Context, node string) (*model.EdgeConfig, error) {
	var cfg model.EdgeConfig
	err := DB.GetContext(ctx, &cfg, fmt.Sprintf(
		`SELECT * from %s where node_id = ?`, tableEdgeConfig), node)
	return &cfg, err
}

func SetEdgeConfig(ctx context.Context, cfg *model.EdgeConfig) error {

	var oldCfg model.EdgeConfig
	err := DB.GetContext(ctx, &oldCfg, fmt.Sprintf(
		`SELECT * from %s where node_id = ?`, tableEdgeConfig), cfg.NodeId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		_, err = DB.NamedExecContext(ctx, fmt.Sprintf(
			`INSERT INTO %s (node_id, config, created_at, updated_at) values (:node_id, :config, :created_at, :updated_at)`, tableEdgeConfig), cfg)
	}
	if err == nil {
		_, err = DB.NamedExecContext(ctx, fmt.Sprintf(
			`UPDATE %s SET config = :config, updated_at = :updated_at where node_id = :node_id`, tableEdgeConfig), cfg)
	}

	return err
}
