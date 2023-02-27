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

const maxPageSize = 500

type NodeFetcher struct {
	BaseFetcher
}

func newNodeFetcher() *NodeFetcher {
	return &NodeFetcher{BaseFetcher: newBaseFetcher()}
}

func (n *NodeFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start to fetch all nodes")
	start := time.Now()
	defer func() {
		log.Infof("fetch all nodes done, cost: %v", time.Since(start))
	}()

	var total int64
	page, size := 1, maxPageSize

loop:
	offset := (page - 1) * size
	resp, err := scheduler.Api.ListNodes(ctx, offset, size)
	if err != nil {
		log.Errorf("api ListNodes: %v", err)
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

	n.Push(ctx, func() error {
		err = dao.BulkUpsertDeviceInfo(ctx, nodes)
		if err != nil {
			log.Errorf("bulk upsert device info: %v", err)
		}

		if err = addDeviceInfoHours(ctx, nodes); err != nil {
			log.Errorf("add device info hours: %v", err)
		}
		return nil
	})

	if total < resp.Total {
		goto loop
	}

	// add inactive node records for statistics
	err = dao.GenerateInactiveNodeRecords(context.Background(), start)
	if err != nil {
		log.Errorf("generate inactive node records: %v", err)
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
		deviceInfo.IpProvince = ipLocationList[1]
		deviceInfo.IpCity = ipLocationList[len(ipLocationList)-1]
	}
	deviceInfo.CpuUsage = utils.ToFixed(deviceInfo.CpuUsage, 2)
	deviceInfo.MemoryUsage = utils.ToFixed(deviceInfo.MemoryUsage, 2)
	deviceInfo.BandwidthUp = utils.ToFixed(deviceInfo.BandwidthUp/gigaBytes, 2)
	deviceInfo.BandwidthDown = utils.ToFixed(deviceInfo.BandwidthDown/gigaBytes, 2)
	deviceInfo.TotalUpload = utils.ToFixed(deviceInfo.TotalUpload/gigaBytes, 2)
	deviceInfo.TotalDownload = utils.ToFixed(deviceInfo.TotalDownload/gigaBytes, 2)
	deviceInfo.DiskSpace = utils.ToFixed(deviceInfo.DiskSpace/teraBytes, 4)
	deviceInfo.DiskUsage = utils.ToFixed(deviceInfo.DiskUsage, 2)
	deviceInfo.ActiveStatus = 1
	return &deviceInfo
}

var _ Fetcher = &NodeFetcher{}
