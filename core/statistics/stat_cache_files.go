package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

const DKeyCacheFiles = "dk_cache_files"

func (s *Statistic) StatCacheFilesMinutes() {
	ctx := context.Background()
	s.Once(ctx, DKeyCacheFiles, func() error {
		log.Info("start get cache files info")
		start := time.Now()
		defer func() {
			log.Infof("get cache files info done, cost: %v", time.Since(start))
		}()

		fullNodeInfoHour := &model.FullNodeInfoHour{}
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

		fullNodeInfoHour.CarFileCount = int64(resp.CarFileCount)
		fullNodeInfoHour.FileDownloadCount = int64(resp.DownloadCount)
		fullNodeInfoHour.TotalFileSize = float64(resp.TotalSize)
		fullNodeInfoHour.ValidatorCount = int32(state.AllVerifier)
		fullNodeInfoHour.CandidateCount = int32(state.AllCandidate)
		fullNodeInfoHour.EdgeCount = int32(state.AllEdgeNode)
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
	})
}
