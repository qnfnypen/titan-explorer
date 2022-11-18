package core

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/query"
	"github.com/gnasnik/titan-explorer/core/oplog"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) error {
	if cfg.DatabaseURL == "" {
		return errors.New("database url not setup")
	}

	db, err := gorm.Open(mysql.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return err
	}

	query.SetDefault(db)
	dao.DB = db
	oplog.Subscribe(context.Background())
	return nil
}
