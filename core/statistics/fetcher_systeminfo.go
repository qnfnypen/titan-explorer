package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

type SystemInfoFetcher struct {
	BaseFetcher
}

func newSystemInfoFetcher() *SystemInfoFetcher {
	return &SystemInfoFetcher{BaseFetcher: newBaseFetcher()}
}

func (s *SystemInfoFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start to fetch full node asset count info")
	start := time.Now()
	defer func() {
		log.Infof("count fetch system info done, cost: %v", time.Since(start))
	}()

	respFromValidationInfo, err := scheduler.Api.GetValidationInfo(ctx)
	if err != nil {
		log.Errorf("api GetSystemInfo: %v", err)
		return err
	}

	intCount, err := scheduler.Api.GetAssetCount(ctx)
	if err != nil {
		log.Errorf("api GetAssetCount: %v", err)
		return err
	}
	if intCount > 0 {
		config.GNodesInfo.AssetCount = int64(intCount)
	}
	s.Push(ctx, func() error {
		if err := dao.UpsertSystemInfo(ctx, &model.SystemInfo{
			SchedulerUuid: scheduler.Uuid,
			// todo
			CarFileCount:     0,
			DownloadCount:    0,
			NextElectionTime: respFromValidationInfo.NextElectionTime,
		}); err != nil {
			log.Errorf("upsert system info: %v", err)
		}
		return nil
	})

	return nil
}

var _ Fetcher = &SystemInfoFetcher{}
