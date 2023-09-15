package statistics

import (
	"context"
	"database/sql"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"time"
)

const (
	defaultRequestLimit = 500
	defaultBackupDays   = 7
)

type AssertFetcher struct {
	BaseFetcher
}

func (a AssertFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start to fetch 【assert info】")
	start := time.Now()
	defer func() {
		log.Infof("fetch assert files, cost: %v", time.Since(start))
	}()

	latest, err := dao.GetLatestAsset(ctx)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	var (
		startTime, endTime time.Time
		limit, offset      int
	)

	if latest == nil || latest.EndTime.IsZero() {
		startTime = carbon.Now().StartOfDay().SubDays(defaultBackupDays).Carbon2Time()
	} else {
		startTime = latest.EndTime
	}

	limit = defaultRequestLimit
	endTime = carbon.Now().EndOfDay().Carbon2Time()

Loop:
	assertsRes, err := scheduler.Api.GetReplicaEvents(ctx, startTime, endTime, limit, offset)
	if err != nil {
		log.Errorf("client api GetReplicaEvents: %v", err)
		return err
	}

	offset += len(assertsRes.ReplicaEvents)

	asserts, err := toAssets(assertsRes.ReplicaEvents)
	if err != nil {
		log.Errorf("toAssets: %v", err)
		return err
	}

	err = dao.AddAssets(ctx, asserts)
	if err != nil {
		log.Errorf("create user info hour: %v", err)
	}

	if assertsRes.Total > offset {
		goto Loop
	}

	return nil
}

func toAssets(in []*types.ReplicaEventInfo) ([]*model.Asset, error) {
	var out []*model.Asset
	for _, r := range in {
		out = append(out, &model.Asset{
			NodeID:     r.NodeID,
			Event:      int64(r.Event),
			Cid:        r.Cid,
			Hash:       r.Hash,
			TotalSize:  r.TotalSize,
			Expiration: r.Expiration,
			EndTime:    r.EndTime,
		})
	}
	return out, nil
}

func newAssertFetcher() *AssertFetcher {
	return &AssertFetcher{BaseFetcher: newBaseFetcher()}
}

var _ Fetcher = &AssertFetcher{}
