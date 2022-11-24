package dao

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"strings"
)

var (
	// DB reference to database
	DB *sqlx.DB
)

type QueryOption struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	Order      string `json:"order" form:"order"`
	OrderField string `json:"order_field" form:"order_field"`
	StartTime  string `json:"startTime" form:"start_time"`
	EndTime    string `json:"endTime" form:"end_time"`
	UserID     string `json:"userId" form:"user_id"`
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
