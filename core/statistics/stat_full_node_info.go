package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

func (s *Statistic) StatFullNodeInfoByMinutes() error {
	log.Info("start state full node info")
	start := time.Now()
	defer func() {
		log.Infof("state full node done, cost: %v", time.Since(start))
	}()

	fullNodeInfoHour := &model.FullNodeInfoHour{}
	ctx := context.Background()
	resp, err := s.api.StatCaches(ctx)
	if err != nil {
		log.Errorf("stat caches: %v", err)
		return err
	}

	state, err := s.api.StateNetwork(ctx)
	if err != nil {
		log.Errorf("state network: %v", err)
		return err
	}

	fullNodeInfoHour.TotalCarfile = int64(resp.CarFileCount)
	fullNodeInfoHour.DownloadCount = int64(resp.DownloadCount)
	fullNodeInfoHour.TotalCarfileSize = float64(resp.TotalSize)
	fullNodeInfoHour.ValidatorCount = int32(state.AllVerifier)
	fullNodeInfoHour.CandidateCount = int32(state.AllCandidate)
	fullNodeInfoHour.EdgeCount = int32(state.AllEdgeNode)
	fullNodeInfoHour.TotalNodeCount = int32(state.AllVerifier + state.AllEdgeNode + state.AllCandidate)
	fullNodeInfoHour.TotalDownloadBandwidth = state.TotalBandwidthDown
	fullNodeInfoHour.TotalUplinkBandwidth = state.TotalBandwidthUp
	fullNodeInfoHour.TotalStorage = state.StorageT
	fullNodeInfoHour.Time = time.Now()
	fullNodeInfoHour.CreatedAt = time.Now()
	err = dao.AddFullNodeInfoHours(ctx, fullNodeInfoHour)
	if err != nil {
		log.Errorf("add full node info hours: %v", err)
		return err
	}

	return nil
}
