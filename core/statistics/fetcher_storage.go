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
	//if resp.Total <= 0 {
	//	go c.FetchFromToken(ctx, scheduler)
	//	return nil
	//}
	//var events []*model.RetrievalEvent
	//for _, data := range resp.WorkloadRecordInfos {
	//	Workload := decodeWorkload(data.ClientWorkload)
	//	events = append(events, toRetrievalEvent(Workload, data))
	//
	//}
	//sum += len(resp.WorkloadRecordInfos)
	//offset++
	//req.Cursor += len(resp.AssetEventInfos)

	//log.Debugf("GetCacheTaskInfos got %d/%d blocks", sum, resp.Total)
	//
	//c.Push(ctx, func() error {
	//	err = dao.CreateRetrievalEvents(ctx, events)
	//	if err != nil {
	//		log.Errorf("create cacheEvents: %v", err)
	//	}
	//	go toUpdateDownloadCount(ctx, events)
	//	return nil
	//})
	//if sum < resp.Total {
	//	goto loop
	//}
	return nil
}

//
//func (c *RetrievalFetcher) FetchFromToken(ctx context.Context, scheduler *Scheduler) {
//	log.Info("start to fetch retrieval files unhandled")
//	start := time.Now()
//	defer func() {
//		log.Infof("fetch retrieval files unhandled, cost: %v", time.Since(start))
//	}()
//	tokenIds, err := dao.GetUnfinishedEvent(ctx)
//	if err != nil {
//		log.Errorf("get token id from retrieval event: %v", err)
//		return
//	}
//	if tokenIds == nil {
//		return
//	}
//	var events []*model.RetrievalEvent
//	for _, tokenId := range tokenIds {
//		resp, err := scheduler.Api.GetWorkloadRecord(ctx, tokenId)
//		if err != nil {
//			log.Errorf("client api GetWorkloadRecord: %v", err)
//			return
//		}
//		if resp.Status == 0 {
//			continue
//		}
//		Workload := decodeWorkload(resp.ClientWorkload)
//		events = append(events, toRetrievalEvent(Workload, resp))
//	}
//	if len(events) == 0 {
//		return
//	}
//	c.Push(ctx, func() error {
//		err = dao.CreateRetrievalEvents(ctx, events)
//		if err != nil {
//			log.Errorf("create cacheEvents: %v", err)
//		}
//		//go toUpdateDownloadCount(ctx, events)
//		return nil
//	})
//	return
//}
//
//func decodeWorkload(workLoad []byte) *types.Workload {
//	cWorkload := &types.Workload{}
//	if len(workLoad) > 0 {
//		dec := gob.NewDecoder(bytes.NewBuffer(workLoad))
//		err := dec.Decode(cWorkload)
//		if err != nil {
//			log.Errorf("decode data to *types.Workload error: %w", err)
//
//		}
//	}
//	return cWorkload
//}
//
//func toRetrievalEvent(data *types.Workload, TokenPayload *types.WorkloadRecord) *model.RetrievalEvent {
//	timeStr := time.Unix(data.EndTime, 0)
//	if TokenPayload.Status != 1 {
//		timeStr = time.Now()
//	}
//	return &model.RetrievalEvent{
//		DeviceID: TokenPayload.NodeID,
//		//TokenID:    TokenPayload.ID,
//		ClientID:   TokenPayload.ClientID,
//		CarfileCid: TokenPayload.AssetCID,
//		// todo blocks is null
//		Time:   timeStr,
//		Status: int32(TokenPayload.Status),
//		//Blocks:    data.BlockCount,
//		BlockSize: float64(data.DownloadSize),
//		StartTime: data.StartTime,
//		EndTime:   data.EndTime,
//	}
//}
//
//func toUpdateDownloadCount(ctx context.Context, Events []*model.RetrievalEvent) {
//	for _, Event := range Events {
//		if Event.Status == 1 {
//			// handle client download data
//			err := dao.CountUploadTraffic(ctx, Event.ClientID)
//			if err != nil {
//				log.Errorf("CountUploadTraffic err:%v", err)
//			}
//		}
//		err := dao.CountRetrievalEvent(ctx, Event.DeviceID)
//		if err != nil {
//			log.Errorf("update device info from event: %v", err)
//		}
//	}
//}