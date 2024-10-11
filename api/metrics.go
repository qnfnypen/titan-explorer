package api

import (
	"context"
	"time"

	"github.com/gnasnik/titan-explorer/core/dao"
)

var (
	lockkey  = "gather_lock_key"
	locktime = 30 * time.Second

	updateInterval = 60 * time.Second
)

func SetPrometheusGatherer(ctx context.Context) {
	// 仅在一个副本里运行
	if f, err := dao.AcquireLock(dao.RedisCache, lockkey, "1", locktime); err != nil {
		log.Errorf("[db get lock failed], %s", err.Error())
		return
	} else {
		//	未获取到不执行
		if !f {
			return
		}
	}

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Error("[metrics gatherer] context cancelled.")
			return
		case <-ticker.C:
			log.Info("[metrics gatherer] updating storage prometheus view")
			setStorageGatherer(ctx)
			log.Info("[metrics gatherer] updating storage l1 view")
			setL1Gatherer(ctx)
		}
	}

}
