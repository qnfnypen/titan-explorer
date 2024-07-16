package statistics

import (
	"context"
)

type StorageFetcher struct {
	BaseFetcher
}

func init() {
	// Register newStorageFetcher during initialization
	RegisterFetcher(newStorageFetcher)
}

func newStorageFetcher() Fetcher {
	return &StorageFetcher{BaseFetcher: newBaseFetcher()}
}

var _ Fetcher = &StorageFetcher{}

// Fetch fetches storage information and processes the data.
func (c *StorageFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	// log.Info("start fetching storage info")
	// start := time.Now()
	// defer func() {
	// 	log.Infof("fetched cache files, cost: %v", time.Since(start))
	// }()

	// userIds, err := dao.GetUserIds(ctx)
	// if err != nil {
	// 	log.Errorf("failed to get UserIds: %v", err)
	// 	return err
	// }

	// infos, err := scheduler.Api.GetUserInfos(ctx, userIds)
	// if err != nil {
	// 	log.Errorf("failed to fetch user infos from API: %v", err)
	// 	return err
	// }

	// var mus []*model.UserInfo
	// for userId, user := range infos {
	// 	mus = append(mus, &model.UserInfo{
	// 		UserId:         userId,
	// 		TotalSize:      user.TotalSize,
	// 		UsedSize:       user.UsedSize,
	// 		TotalBandwidth: user.TotalTraffic,
	// 		PeakBandwidth:  user.PeakBandwidth,
	// 		DownloadCount:  user.DownloadCount,
	// 		Time:           start,
	// 		CreatedAt:      time.Now(),
	// 		UpdatedAt:      time.Now(),
	// 	})
	// }

	// if len(mus) == 0 {
	// 	return nil
	// }

	// err = dao.BulkUpsertStorageHours(ctx, mus)
	// if err != nil {
	// 	log.Errorf("failed to create user info hour: %v", err)
	// }

	return nil
}

func (c *StorageFetcher) Finalize() error {
	return nil
}
