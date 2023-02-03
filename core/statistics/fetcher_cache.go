package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"github.com/ipfs/go-cid"
	"github.com/linguohua/titan/api"
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
	log.Info("start to fetch cache files")
	start := time.Now()
	defer func() {
		log.Infof("fetch cache files, cost: %v", time.Since(start))
	}()

	var (
		startTime, endTime time.Time
		sum                int64
	)

	lastEvent, err := dao.GetLastCacheEvent(ctx)
	if err != nil {
		log.Errorf("get last cache event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime = carbon.Now().StartOfMonth().StartOfMinute().Carbon2Time()
	} else {
		startTime = lastEvent.CreatedAt.Add(time.Second)
	}

	endTime = carbon.Time2Carbon(start).SubMinutes(start.Minute() % 5).StartOfMinute().Carbon2Time()
	req := api.ListCacheInfosReq{
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Cursor:    0,
		Count:     maxPageSize,
	}

loop:
	resp, err := scheduler.Api.GetCacheTaskInfos(ctx, req)
	if err != nil {
		log.Errorf("client api GetCacheTaskInfos: %v", err)
		return err
	}

	if resp.Total <= 0 {
		return nil
	}

	var events []*model.CacheEvent
	for _, data := range resp.Datas {
		events = append(events, toCacheEvent(data))
	}

	sum += int64(len(resp.Datas))
	req.Cursor += len(resp.Datas)

	log.Debugf("GetCacheTaskInfos got %d/%d blocks", sum, resp.Total)

	c.Push(ctx, func() error {
		err = dao.CreateCacheEvents(ctx, events)
		if err != nil {
			log.Errorf("create cacheEvents: %v", err)
		}
		return nil
	})

	if sum < resp.Total {
		goto loop
	}

	return nil
}

var _ Fetcher = &CacheFetcher{}

func toValidationEvent(in api.ValidateResultInfo) *model.ValidationEvent {
	return &model.ValidationEvent{
		DeviceID:        in.DeviceID,
		ValidatorID:     in.ValidatorID,
		Status:          int32(in.Status),
		Blocks:          in.BlockNumber,
		Time:            in.ValidateTime,
		Duration:        in.Duration,
		UpstreamTraffic: utils.ToFixed(in.UploadTraffic/megaBytes, 2),
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

func toCacheEvent(data *api.CacheTaskInfo) *model.CacheEvent {
	return &model.CacheEvent{
		DeviceID:   data.DeviceID,
		CarfileCid: hashToCID(data.CarfileHash),
		Blocks:     int64(data.DoneBlocks),
		BlockSize:  float64(data.DoneSize),
		Time:       data.CreateTime,
	}
}
