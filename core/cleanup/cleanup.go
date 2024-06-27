package cleanup

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/golang-module/carbon/v2"
	logging "github.com/ipfs/go-log/v2"
	"time"
)

var log = logging.Logger("cleanup")

var (
	cleanupInterval = time.Minute * 60
)

func Run(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	var isRunning bool

	for {
		select {
		case <-ticker.C:
			if isRunning {
				continue
			}

			isRunning = true

			deleteFrom := carbon.Now().SubDays(2).StdTime()
			if err := cleanUpDeviceInfoHour(ctx, deleteFrom); err != nil {
				log.Errorf("cleanUpDeviceInfoHour: %v", err)
			}

			isRunning = false

		case <-ctx.Done():
		}
	}
}

func cleanUpDeviceInfoHour(ctx context.Context, before time.Time) error {
	query := "select ifnull(max(id),0) from device_info_hour where created_at < ?"

	var maxId int64
	err := dao.DB.GetContext(ctx, &maxId, query, before)
	if err != nil {
		return err
	}

	for {
		deleteSql := `delete from device_info_hour where id <= ? limit 100000`

		res, err := dao.DB.ExecContext(ctx, deleteSql, maxId)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows <= 0 {
			return nil
		}

		log.Infof("deleted  device info hour before %v rows: %d", before, rows)
	}
}
