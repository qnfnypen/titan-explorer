package core

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/oplog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

const (
	maxOpenConnections = 60
	connMaxLifetime    = 120
	maxIdleConnections = 30
	connMaxIdleTime    = 20
)

func Init(cfg *config.Config) error {
	if cfg.DatabaseURL == "" {
		return errors.New("database url not setup")
	}

	db, err := sqlx.Connect("mysql", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)

	dao.DB = db
	oplog.Subscribe(context.Background())
	return nil
}
