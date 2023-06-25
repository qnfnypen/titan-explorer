package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

type ValidationFetcher struct {
	BaseFetcher
}

func newValidationFetcher() *ValidationFetcher {
	return &ValidationFetcher{BaseFetcher: newBaseFetcher()}
}

//func (v *ValidationFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
//	log.Info("start to fetch 【validation events】")
//	start := time.Now()
//	defer func() {
//		log.Infof("fetch validation events done, cost: %v", time.Since(start))
//	}()
//
//	var (
//		startTime, endTime time.Time
//		sum                int64
//		page, pageSize     = 1, maxPageSize
//	)
//
//	lastEvent, err := dao.GetLastValidationEvent(ctx)
//	if err != nil {
//		log.Errorf("get last validation event: %v", err)
//		return err
//	}
//
//	if lastEvent == nil {
//		startTime, _ = time.Parse(utils.TimeFormatDateOnly, utils.TimeFormatDateOnly)
//	} else {
//		startTime = floorFiveMinute(lastEvent.Time)
//	}
//
//	endTime = floorFiveMinute(time.Now())
//
//loop:
//	resp, err := scheduler.Api.GetValidationResults(ctx, startTime, endTime, page, pageSize)
//	if err != nil {
//		log.Errorf("api GetSummaryValidateMessage: %v", err)
//		return err
//	}
//	if resp.Total <= 0 {
//		return nil
//	}
//
//	sum += int64(len(resp.ValidationResultInfos))
//	page++
//
//	var validationEvents []*model.ValidationEvent
//	for _, blockInfo := range resp.ValidationResultInfos {
//		validationEvents = append(validationEvents, toValidationEvent(blockInfo))
//	}
//
//	log.Debugf("GetSummaryValidateMessage got %d/%d messages", sum, resp.Total)
//
//	v.Push(ctx, func() error {
//		err = dao.CreateValidationEvent(ctx, validationEvents)
//		if err != nil {
//			log.Errorf("create validation events: %v", err)
//		}
//		//go toUpdateValidateDownloadCount(ctx, validationEvents)
//		return nil
//	})
//	if sum < int64(resp.Total) {
//		goto loop
//	}
//
//	return nil
//}

//var _ Fetcher = &ValidationFetcher{}

func toUpdateValidateDownloadCount(ctx context.Context, Events []*model.ValidationEvent) {
	for _, Event := range Events {
		if Event.Status == 1 {
			// handle validator download data
			err := dao.CountUploadTraffic(ctx, Event.ValidatorID)
			if err != nil {
				log.Errorf("CountUploadTraffic err:%v", err)
			}
			// handle node upload data
			err = dao.CountValidateEvent(ctx, Event.DeviceID)
			if err != nil {
				log.Errorf("CountUploadTraffic err:%v", err)
			}
		}
	}
}
