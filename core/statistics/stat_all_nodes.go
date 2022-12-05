package statistics

import (
	"context"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"strings"
	"time"
)

func (s *Statistic) FetchAllNodes() error {
	log.Info("start fetch all nodes")
	start := time.Now()
	defer func() {
		log.Infof("fetch all nodes done, cost: %v", time.Since(start))
	}()

	var total int64
	page, size := 1, 50

loop:
	offset := (page - 1) * size
	ctx := context.Background()
	resp, err := s.api.ListNodes(ctx, offset, size)
	if err != nil {
		return err
	}

	total += int64(len(resp.Data))
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

	if err = addDeviceInfoHours(ctx, nodes); err != nil {
		log.Errorf("add device info hours: %v", err)
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
	ipLocationList := strings.Split(deviceInfo.IpLocation, "-")
	if len(ipLocationList) >= 2 {
		deviceInfo.IpCountry = ipLocationList[0]
		deviceInfo.IpCity = ipLocationList[len(ipLocationList)-1]
	}
	return &deviceInfo
}

func (s *Statistic) CountFullNodeInfo() error {
	log.Info("start count full node info")
	start := time.Now()
	defer func() {
		log.Infof("count full node done, cost: %v", time.Since(start))
	}()

	ctx := context.Background()
	resp, err := s.api.StatCaches(ctx)
	if err != nil {
		log.Errorf("stat caches: %v", err)
		return err
	}

	fullNodeInfoHour, err := dao.CountFullNodeInfo(ctx)
	if err != nil {
		log.Errorf("count full node: %v", err)
		return err
	}

	fullNodeInfoHour.TotalCarfile = int64(resp.CarFileCount)
	fullNodeInfoHour.RetrievalCount = int64(resp.DownloadCount)
	fullNodeInfoHour.TotalCarfileSize = float64(resp.TotalSize)

	fullNodeInfoHour.Time = time.Now()
	fullNodeInfoHour.CreatedAt = time.Now()
	err = dao.AddFullNodeInfoHours(ctx, fullNodeInfoHour)
	if err != nil {
		log.Errorf("add full node info hours: %v", err)
		return err
	}

	return nil
}
