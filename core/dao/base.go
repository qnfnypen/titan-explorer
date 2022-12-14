package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/go-redis/redis/v9"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

var (
	// DB reference to database
	DB *sqlx.DB
	// Cache  redis caching instance
	Cache *redis.Client
)

const (
	maxOpenConnections = 60
	connMaxLifetime    = 120
	maxIdleConnections = 30
	connMaxIdleTime    = 20
)

func Init(cfg *config.Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database url not setup")
	}

	db, err := sqlx.Connect("mysql", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	DB = db
	Cache = client
	return nil
}

type QueryOption struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Order      string `json:"order"`
	OrderField string `json:"order_field"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time" `
	UserID     string `json:"user_id"`
}

func GetQueryDataList(sqlClause string, args ...interface{}) ([]map[string]string, error) {
	rows, err := DB.Query(sqlClause, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	dataList := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		data := make(map[string]string)
		for i, col := range values {
			//			if col == nil {
			//				continue
			//			}

			key := columns[i]
			key = strings.ToLower(key)
			data[key] = string(col)

		}
		//		log.Info(&data)
		dataList = append(dataList, data)
	}

	return dataList, nil
}
