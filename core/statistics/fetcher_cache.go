package statistics

import (
	"context"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"math"
	"time"
)

type CacheFetcher struct {
	BaseFetcher
}

func newCacheFetcher() *CacheFetcher {
	return &CacheFetcher{BaseFetcher: newBaseFetcher()}
}

func (c *CacheFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start to fetch 【cache events】")
	start := time.Now()
	defer func() {
		log.Infof("fetch cache files, cost: %v", time.Since(start))
	}()

	var (
		startTime, endTime time.Time
		sum                int
	)

	lastEvent, err := dao.GetLastCacheEvent(ctx)
	if err != nil {
		log.Errorf("get last cache event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime, _ = time.Parse(utils.TimeFormatDateOnly, utils.TimeFormatDateOnly)
	} else {
		startTime = lastEvent.Time.Add(time.Second)
	}

	endTime = carbon.Time2Carbon(start).SubMinutes(start.Minute() % 5).StartOfMinute().Carbon2Time()
	size, offset := maxPageSize, 1
loop:
	resp, err := scheduler.Api.GetAssetEvents(ctx, startTime, endTime, size, (offset-1)*size)
	if err != nil {
		log.Errorf("client api GetCacheTaskInfos: %v", err)
		return err
	}

	if resp.Total <= 0 {
		return nil
	}
	var events []*model.CacheEvent
	for _, data := range resp.AssetEventInfos {
		eventCid := hashToCID(data.Hash)
		respRecord, err := scheduler.Api.GetAssetRecord(ctx, eventCid)
		if err != nil {
			log.Errorf("client api GetAssetRecord: %v", err)
			return err
		}
		err = dao.ResetCacheEvents(ctx, eventCid)
		if err != nil {
			log.Errorf("client api ResetCacheEvents: %v", err)
			return err
		}
		lenReplicaInfo := len(respRecord.ReplicaInfos)
		if lenReplicaInfo == 0 {
			var nilType types.ReplicaInfo
			nilType.EndTime = respRecord.EndTime
			nilType.Status = 4
			event := toCacheEvent(respRecord, &nilType, int32(lenReplicaInfo))
			events = append(events, event)
		} else {
			for _, ReplicaInfo := range respRecord.ReplicaInfos {
				event := toCacheEvent(respRecord, ReplicaInfo, int32(lenReplicaInfo))
				events = append(events, event)
			}
		}

	}
	sum += len(resp.AssetEventInfos)
	offset++
	//req.Cursor += len(resp.AssetEventInfos)

	log.Debugf("cacheEvents got %d/%d AssetEventInfos", sum, resp.Total)
	if len(events) == 0 {
		return nil
	}
	c.Push(ctx, func() error {
		err = dao.CreateCacheEvents(ctx, events)
		if err != nil {
			log.Errorf("create cacheEvents: %v", err)
		}
		go toUpdateDeviceInfo(ctx, events)
		return nil
	})

	if sum < resp.Total {
		goto loop
	}

	return nil
}

var _ Fetcher = &CacheFetcher{}

func toValidationEvent(in types.ValidationResultInfo) *model.ValidationEvent {
	return &model.ValidationEvent{
		DeviceID:        in.NodeID,
		ValidatorID:     in.ValidatorID,
		Status:          int32(in.Status),
		Blocks:          in.BlockNumber,
		Time:            in.StartTime,
		Duration:        in.Duration,
		UpstreamTraffic: utils.ToFixed(float64(in.Duration)*in.Bandwidth, 2),
	}
}

func hashToCID(hashString string) string {
	multiphase, err := mh.FromHexString(hashString)
	if err != nil {
		return ""
	}
	cid := cid.NewCidV1(cid.Raw, multiphase)
	return cid.String()
}

func (s *Statistic) CountRetrievals() error {
	log.Info("start to count retrievals")
	start := time.Now()
	defer func() {
		log.Infof("count retrievals, cost: %v", time.Since(start))
	}()

	var startTime time.Time
	ctx := context.Background()
	lastEvent, err := dao.GetLastRetrievalEvent(ctx)
	if err != nil {
		log.Errorf("get last retrieval event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime = carbon.Now().SubDays(60).Carbon2Time()
	} else {
		startTime = floorFiveMinute(lastEvent.Time)
	}

	now := time.Now()
	oneDay := 24 * time.Hour
	for st := startTime; st.Before(now); {
		startT := st
		st = st.Add(oneDay)
		endT := st
		events, err := dao.GenerateRetrievalEvents(ctx, startT, endT)
		if err != nil {
			log.Errorf("generate retrieval events: %v", err)
			continue
		}

		if len(events) == 0 {
			continue
		}

		err = dao.CreateRetrievalEvents(ctx, events)
		if err != nil {
			log.Errorf("create retrieve events %v", err)
			continue
		}
	}

	return nil
}

func floorFiveMinute(t time.Time) time.Time {
	year, month, day := t.Date()
	hour := t.Hour()
	minute := int(5 * (math.Floor(float64(t.Minute() / 5))))
	return time.Date(year, month, day, hour, minute, 0, 0, time.Local)
}

func toCacheEvent(assetRecord *types.AssetRecord, data *types.ReplicaInfo, lenReplicaInfo int32) *model.CacheEvent {
	return &model.CacheEvent{
		DeviceID:     data.NodeID,
		CarfileCid:   hashToCID(assetRecord.Hash),
		ReplicaInfos: lenReplicaInfo,
		Blocks:       assetRecord.TotalBlocks,
		BlockSize:    float64(assetRecord.TotalSize),
		Time:         data.EndTime,
		// todo file create time
		Status:    int32(data.Status),
		UpdatedAt: time.Now(),
	}
}

func toUpdateDeviceInfo(ctx context.Context, Events []*model.CacheEvent) {
	for _, Event := range Events {
		err := dao.CountCacheEvent(ctx, Event.DeviceID)
		if err != nil {
			log.Errorf("update device info from event: %v", err)
		}
	}
}
