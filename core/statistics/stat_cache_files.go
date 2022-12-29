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
	"sync"
	"time"
)

func (s *Statistic) asyncExecute(jobs []func() error) {
	var wg sync.WaitGroup
	wg.Add(len(jobs))

	for _, job := range jobs {
		go func(job func() error) {
			defer wg.Done()
			if err := job(); err != nil {
				log.Errorf("handling job: %v", err)
			}
		}(job)
	}

	wg.Wait()
	return
}

func (s *Statistic) CountCacheFiles() error {
	log.Info("start count cache files")
	start := time.Now()
	defer func() {
		log.Infof("count cache files, cost: %v", time.Since(start))
	}()

	var (
		startTime, endTime time.Time
		sum                int64
	)

	ctx := context.Background()
	lastEvent, err := dao.GetLastCacheEvent(ctx)
	if err != nil {
		log.Errorf("get last cache event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime = carbon.Now().StartOfDay().StartOfMinute().Carbon2Time()
	} else {
		startTime = lastEvent.MaxCreatedTime.Add(time.Second)
	}

	endTime = carbon.Time2Carbon(start).SubMinutes(start.Minute() % 5).StartOfMinute().Carbon2Time()
	req := api.ListCacheBlocksReq{
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Cursor:    0,
		Count:     500,
	}

loop:
	resp, err := s.api.GetCacheBlockInfos(ctx, req)
	if err != nil {
		log.Errorf("api GetCacheBlockInfos: %v", err)
		return err
	}

	var blockInfos []*model.BlockInfo
	for _, blockInfo := range resp.Data {
		blockInfos = append(blockInfos, toBlockInfo(blockInfo))
	}

	if resp.Total <= 0 {
		return nil
	}

	sum += int64(len(resp.Data))
	req.Cursor += len(resp.Data)

	log.Debugf("GetCacheBlockInfos got %d/%d blocks", sum, resp.Total)

	err = dao.CreateBlockInfo(ctx, blockInfos)
	if err != nil {
		log.Errorf("create block info: %v", err)
	}

	if sum < resp.Total {
		<-time.After(100 * time.Millisecond)
		goto loop
	}

	err = dao.TxStatisticDeviceBlocks(ctx, startTime, endTime)
	if err != nil {
		log.Errorf("statistics device blocks: %v", err)
		return err
	}

	return nil
}

func toBlockInfo(in api.BlockInfo) *model.BlockInfo {
	return &model.BlockInfo{
		DeviceID:    in.DeviceID,
		CarfileHash: in.CarfileHash,
		CarfileCid:  hashToCID(in.CarfileHash),
		Status:      int32(in.Status),
		Size:        utils.ToFixed(float64(in.Size)/gibiByte, 2),
		CreatedTime: in.CreateTime,
		EndTime:     in.EndTime,
	}
}

func toValidationEvent(in api.ValidateResultInfo) *model.ValidationEvent {
	return &model.ValidationEvent{
		DeviceID:        in.DeviceID,
		ValidatorID:     in.ValidatorID,
		Status:          int32(in.Status),
		Blocks:          in.BlockNumber,
		Time:            in.ValidateTime,
		Duration:        in.Duration,
		UpstreamTraffic: utils.ToFixed(in.UploadTraffic/gibiByte, 2),
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

func (s *Statistic) FetchValidationEvents() error {
	log.Info("start fetch validation events")
	start := time.Now()
	defer func() {
		log.Infof("fetch validation events done, cost: %v", time.Since(start))
	}()

	var (
		startTime, endTime time.Time
		sum                int64
		page, pageSize     = 1, 100
	)

	ctx := context.Background()
	lastEvent, err := dao.GetLastValidationEvent(ctx)
	if err != nil {
		log.Errorf("get last validation event: %v", err)
		return err
	}

	if lastEvent == nil {
		startTime = carbon.Now().StartOfDay().StartOfMinute().Carbon2Time()
	} else {
		startTime = lastEvent.Time.Add(time.Second)
	}

	endTime = carbon.Time2Carbon(start).SubMinutes(start.Minute() % 5).StartOfMinute().Carbon2Time()

loop:
	resp, err := s.api.GetSummaryValidateMessage(ctx, startTime, endTime, page, pageSize)
	if err != nil {
		log.Errorf("api GetSummaryValidateMessage: %v", err)
		return err
	}

	if resp.Total <= 0 {
		return nil
	}

	sum += int64(len(resp.ValidateResultInfos))
	page++

	var validationEvents []*model.ValidationEvent
	for _, blockInfo := range resp.ValidateResultInfos {
		validationEvents = append(validationEvents, toValidationEvent(blockInfo))
	}

	log.Debugf("GetSummaryValidateMessage got %d/%d messages", sum, resp.Total)

	err = dao.CreateValidationEvent(ctx, validationEvents)
	if err != nil {
		log.Errorf("create validation events: %v", err)
	}

	if sum < int64(resp.Total) {
		<-time.After(100 * time.Millisecond)
		goto loop
	}

	return nil
}
