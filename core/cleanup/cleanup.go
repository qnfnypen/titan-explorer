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
	cleanupInterval = time.Hour * 24
)

func Run(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			threeMonthAgo := carbon.Now().SubMonths(3).Carbon2Time()
			if err := cleanUpDeviceInfoHour(ctx, threeMonthAgo); err != nil {
				log.Errorf("cleanUpDeviceInfoHour: %v", err)
			}
		case <-ctx.Done():
		}
	}
}

func cleanUpDeviceInfoHour(ctx context.Context, before time.Time) error {
	return dao.DeleteDeviceInfoHourHistory(ctx, before)
}
