package statistics

import (
	"context"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"strings"
	"time"
)

func (s *Statistic) FetchAllNodes() error {
	log.Info("start to fetch all nodes")
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
		log.Errorf("api ListNodes: %v", err)
		return err
	}

	total += int64(len(resp.Data))
	page++

	var nodes []*model.DeviceInfo
	for _, node := range resp.Data {
		if node.DeviceId == "" {
			continue
		}
		nodes = append(nodes, toDeviceInfo(node))
	}

	log.Infof("handling %d/%d nodes", total, resp.Total)

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

	s.asyncExecute(
		[]func() error{
			s.SumDeviceInfoProfit,
			s.CountFullNodeInfo,
			s.CountCacheFiles,
			s.CountRetrievals,
			s.FetchValidationEvents,
		},
	)

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
		deviceInfo.IpProvince = ipLocationList[1]
		deviceInfo.IpCity = ipLocationList[len(ipLocationList)-1]
	}
	deviceInfo.BandwidthUp = utils.ToFixed(deviceInfo.BandwidthUp/gigaBytes, 2)
	deviceInfo.BandwidthDown = utils.ToFixed(deviceInfo.BandwidthDown/gigaBytes, 2)
	deviceInfo.TotalUpload = utils.ToFixed(deviceInfo.TotalUpload/gigaBytes, 2)
	deviceInfo.TotalDownload = utils.ToFixed(deviceInfo.TotalDownload/gigaBytes, 2)
	deviceInfo.DiskSpace = utils.ToFixed(deviceInfo.DiskSpace/teraBytes, 4)
	deviceInfo.ActiveStatus = 1
	return &deviceInfo
}

func (s *Statistic) CountFullNodeInfo() error {
	log.Info("start to count full node info")
	start := time.Now()
	defer func() {
		log.Infof("count full node done, cost: %v", time.Since(start))
	}()

	ctx := context.Background()
	resp, err := s.api.GetSystemInfo(ctx)
	if err != nil {
		log.Errorf("api GetSystemInfo: %v", err)
		return err
	}

	fullNodeInfo, err := dao.CountFullNodeInfo(ctx)
	if err != nil {
		log.Errorf("count full node: %v", err)
		return err
	}

	fullNodeInfo.TotalCarfile = int64(resp.CarFileCount)
	fullNodeInfo.RetrievalCount = int64(resp.DownloadCount)
	fullNodeInfo.NextElectionTime = time.Unix(resp.NextElectionTime, 0)

	fullNodeInfo.Time = time.Now()
	fullNodeInfo.CreatedAt = time.Now()
	err = dao.CacheFullNodeInfo(ctx, fullNodeInfo)
	if err != nil {
		log.Errorf("cache full node info: %v", err)
		return err
	}

	return nil
}
