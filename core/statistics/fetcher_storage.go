package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

type StorageFetcher struct {
	BaseFetcher
}

func newStorageFetcher() *StorageFetcher {
	return &StorageFetcher{BaseFetcher: newBaseFetcher()}
}

var _ Fetcher = &StorageFetcher{}

func (c *StorageFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start to fetch 【storage info】")
	start := time.Now()
	defer func() {
		log.Infof("fetch cache files, cost: %v", time.Since(start))
	}()

	userIds, err := dao.GetUserIds(ctx)
	if err != nil {
		log.Errorf("get GetUserIds: %v", err)
		return err
	}

	infos, err := scheduler.Api.GetUserInfos(ctx, userIds)
	if err != nil {
		log.Errorf("client api GetUserInfos: %v", err)
		return err
	}

	var mus []*model.UserInfo
	for userId, user := range infos {
		mus = append(mus, &model.UserInfo{
			UserId:         userId,
			TotalSize:      user.TotalSize,
			UsedSize:       user.UsedSize,
			TotalBandwidth: user.TotalTraffic,
			PeakBandwidth:  user.PeakBandwidth,
			DownloadCount:  user.DownloadCount,
			Time:           start,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	if len(mus) == 0 {
		return nil
	}

	err = dao.BulkUpsertStorageHours(ctx, mus)
	if err != nil {
		log.Errorf("create user info hour: %v", err)
	}

	return nil
}
