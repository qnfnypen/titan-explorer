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
	var infos []*model.UserInfo
	for _, UserId := range userIds {
		Info, err := scheduler.Api.GetUserInfo(ctx, UserId)
		if err != nil {
			log.Errorf("client api GetUserInfo: %v", err)
			continue
		}
		if Info == nil {
			continue
		}
		var userInfo model.UserInfo
		userInfo.UserId = UserId
		userInfo.TotalSize = Info.TotalSize
		userInfo.UsedSize = Info.UsedSize
		userInfo.TotalBandwidth = Info.TotalTraffic
		userInfo.PeakBandwidth = Info.PeakBandwidth
		userInfo.DownloadCount = Info.DownloadCount
		userInfo.Time = start
		userInfo.CreatedAt = time.Now()
		userInfo.UpdatedAt = time.Now()
		infos = append(infos, &userInfo)
	}
	err = dao.BulkUpsertStorageHours(ctx, infos)
	if err != nil {
		log.Errorf("create user info hour: %v", err)
	}
	return nil
}
