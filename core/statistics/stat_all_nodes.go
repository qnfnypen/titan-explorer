package statistics

import (
	"context"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

func (s *Statistic) FetchAllNodes() error {
	log.Info("start fetch all nodes")
	start := time.Now()
	defer func() {
		log.Infof("fetch all nodes done, cost: %v", time.Since(start))
	}()

	ctx := context.Background()
	var total int64
	page, size := 1, 50

loop:
	resp, err := s.api.ListNodes(ctx, page, size)
	if err != nil {
		return err
	}

	total += resp.Total
	page++

	var nodes []*model.DeviceInfo
	for _, node := range resp.Data {
		nodes = append(nodes, toDeviceInfo(node))
	}

	log.Infof("fetch %d nodes, prepare to update", len(nodes))

	err = dao.BulkUpsertDeviceInfo(ctx, nodes)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}

	if total < resp.Total {
		goto loop
	}

	return nil
}

func toDeviceInfo(v interface{}) *model.DeviceInfo {
	var deviceInfo model.DeviceInfo
	data, err := json.Marshal(v)
	if err != nil {
		log.Errorf("marshal device info: %v", err)
		return nil
	}

	err = json.Unmarshal(data, &deviceInfo)
	if err != nil {
		return nil
	}

	return &deviceInfo
}
