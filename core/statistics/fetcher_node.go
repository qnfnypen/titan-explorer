package statistics

import (
	"context"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/core/geo"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	errs "github.com/pkg/errors"
	"strconv"
	"time"

	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const maxPageSize = 1000

const (
	DeviceStatusOffline  = "offline"
	DeviceStatusOnline   = "online"
	DeviceStatusAbnormal = "abnormal"

	DeviceStatusCodeOffline  = 3
	DeviceStatusCodeOnline   = 1
	DeviceStatusCodeAbnormal = 2
)

// NodeFetcher handles fetching information about all nodes
type NodeFetcher struct {
	BaseFetcher
}

func init() {
	// Register newNodeFetcher during initialization
	RegisterFetcher(newNodeFetcher)
}

// newNodeFetcher creates a new NodeFetcher instance
func newNodeFetcher() Fetcher {
	return &NodeFetcher{BaseFetcher: newBaseFetcher()}
}

// Fetch fetches information about all nodes
// 流程如下:
// 1. 遍历拉取节点的数据, 每次上限为 1000 个(调度器那边设置上限也是1000)
// 2. 区分在线和离线的节点, 创建一个任务, 任务步骤:
// 2.1 更新 device_info表, 使用的是 INSERT INTO ... ON DUPLICATE KEY UPDATE ... , 在线的需要更新多个字段, 离线的只更新在线状态为离线
// 2.2 写入 device_info_hour 表, 每次拉取都会记录到这个表, 5分钟一条记录
// 2.3 统计每个节点当天的 收益,在线等数据, 并写到 device_info_daily 表, 唯一主键为 device_id  和 time, 每个节点每天增加一条记录
// 3. 把任务 Push 到队列等待执行
// 4. Finalize 任务, 执行以下统计
// 4.1 统计每个节点的每日收益,昨日收益,七天收益和月收益等, 更新到 device_info 表
// 4.2 统计所有节点的总收益,总内存和总的存储等总览页面数据的统计
func (n *NodeFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Infof("start fetching all nodes from scheduler: %s", scheduler.AreaId)
	start := time.Now()
	defer func() {
		log.Infof("fetch all nodes done from scheduler: %s, cost: %v", scheduler.AreaId, time.Since(start))
	}()

	userInDevice, err := dao.GetAllDeviceUserIdFromCache(ctx)
	if err != nil {
		log.Errorf("get all device user id from cache: %v", err)
	}

	var total int64
	page, size := 1, maxPageSize

loop:
	offset := (page - 1) * size
	reqStart := time.Now()
	resp, err := scheduler.Api.GetNodeList(ctx, offset, size)
	if err != nil {
		log.Errorf("api GetNodeList from %s: %v", scheduler.AreaId, err)
		return nil
	}
	log.Infof("request GetNodeList from %s cost: %v", scheduler.AreaId, time.Since(reqStart))

	total += int64(len(resp.Data))
	page++

	var (
		onlineNodes     []*model.DeviceInfo
		offlineNodes    []*model.DeviceInfo
		deviceInfoHours []*model.DeviceInfoHour
	)

	for _, node := range resp.Data {
		if node.NodeID == "" {
			continue
		}

		nodeInfo := ToDeviceInfo(node, scheduler.AreaId)
		if nodeInfo.DeviceStatus == DeviceStatusOffline {
			offlineNodes = append(offlineNodes, nodeInfo)
			continue
		}

		userId, ok := userInDevice[nodeInfo.DeviceID]
		if !ok || userId == "" {
			userId = getDeviceUserId(ctx, nodeInfo.DeviceID)
		}

		nodeInfo.UserID = userId
		deviceInfoHours = append(deviceInfoHours, ToDeviceInfoHour(nodeInfo, start))
		onlineNodes = append(onlineNodes, nodeInfo)
	}

	if len(onlineNodes)+len(offlineNodes) < 1 {
		log.Errorf("start to fetch %s all nodes: nodes length is 0", scheduler.AreaId)
		return nil
	}

	log.Infof("handling %s %d/%d nodes, online: %d offline: %d", scheduler.AreaId, total, resp.Total, len(onlineNodes), len(offlineNodes))

	n.Push(ctx, func() error {
		if len(onlineNodes) > 0 {
			err := dao.BulkUpsertDeviceInfo(ctx, onlineNodes)
			if err != nil {
				log.Errorf("%s bulk upsert device info: %v", scheduler.AreaId, err)
			}

			if err = addDeviceInfoHours(ctx, deviceInfoHours); err != nil {
				log.Errorf("add device info hours: %v", err)
			}

			if err := sumDailyReward(ctx, start, onlineNodes); err != nil {
				log.Errorf("add device info daily reward: %v", err)
			}
		}

		if len(offlineNodes) > 0 {
			err = dao.BulkInsertOrUpdateDeviceStatus(ctx, offlineNodes)
			if err != nil {
				log.Errorf("bulk add device info: %v", err)
			}
		}

		return nil
	})

	if total < resp.Total {
		goto loop
	}

	return nil
}

func (n *NodeFetcher) Finalize() error {
	st := time.Now()
	log.Infof("finialize start")
	defer func() {
		log.Infof("finialize done, cost: %v", time.Since(st))
	}()

	if err := n.SumDeviceInfoProfit(); err != nil {
		log.Errorf("sum device info profit: %v", err)
	}

	if err := SumUserDeviceReward2(context.Background()); err != nil {
		log.Errorf("sum user device reward: %v", err)
	}

	if err := SumAllNodes(); err != nil {
		log.Errorf("sum all node: %v", err)
	}
	return nil
}

// sumDailyReward 写入或更新 device_info_daily表
func sumDailyReward(ctx context.Context, sumTime time.Time, devices []*model.DeviceInfo) error {
	log.Infof("start sum daily reward")
	start := time.Now()
	defer func() {
		log.Infof("sum daily reward cost: %v", time.Since(start))
	}()

	end := carbon.Yesterday().EndOfDay().String()

	var deviceIds []string
	for _, device := range devices {
		deviceIds = append(deviceIds, device.DeviceID)
	}

	maxDeviceInfos, err := dao.SumDeviceDailyBeforeDate(ctx, deviceIds, end)
	if err != nil {
		log.Errorf("SumDeviceDailyBeforeDate: %v", err)
		return err
	}

	var updatedDevices []*model.DeviceInfoDaily
	for _, deviceInfo := range devices {
		deviceInfo.UpdatedAt = sumTime
		ud, ok := maxDeviceInfos[deviceInfo.DeviceID]
		if !ok {
			updatedDevices = append(updatedDevices, deviceInfoToDailyInfo(deviceInfo))
			continue
		}

		deviceInfo.CumulativeProfit = deviceInfo.CumulativeProfit - ud.Income
		deviceInfo.OnlineTime = deviceInfo.OnlineTime - ud.OnlineTime
		deviceInfo.UploadTraffic = deviceInfo.UploadTraffic - ud.UpstreamTraffic
		deviceInfo.DownloadTraffic = deviceInfo.DownloadTraffic - ud.DownstreamTraffic
		deviceInfo.RetrievalCount = deviceInfo.RetrievalCount - ud.RetrievalCount
		deviceInfo.CacheCount = deviceInfo.CacheCount - ud.BlockCount

		updatedDevices = append(updatedDevices, deviceInfoToDailyInfo(deviceInfo))
	}

	err = dao.BulkUpsertDeviceInfoDaily(context.Background(), updatedDevices)
	if err != nil {
		return errs.Wrap(err, "bulk upsert device info daily")
	}

	return nil
}

func deviceInfoToDailyInfo(deviceInfo *model.DeviceInfo) *model.DeviceInfoDaily {
	return &model.DeviceInfoDaily{
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		UserID:            deviceInfo.UserID,
		DeviceID:          deviceInfo.DeviceID,
		Time:              carbon.Time2Carbon(deviceInfo.UpdatedAt).StartOfDay().AddHours(8).Carbon2Time(),
		Income:            deviceInfo.CumulativeProfit,
		OnlineTime:        deviceInfo.OnlineTime,
		PkgLossRatio:      0, // todo
		Latency:           0, //todo
		NatRatio:          0,
		DiskUsage:         deviceInfo.DiskUsage,
		DiskSpace:         deviceInfo.DiskSpace,
		BandwidthUp:       deviceInfo.BandwidthUp,
		BandwidthDown:     deviceInfo.BandwidthDown,
		UpstreamTraffic:   deviceInfo.UploadTraffic,
		DownstreamTraffic: deviceInfo.DownloadTraffic,
		RetrievalCount:    deviceInfo.RetrievalCount,
		BlockCount:        deviceInfo.CacheCount,
	}
}

func ToDeviceInfoHour(device *model.DeviceInfo, t time.Time) *model.DeviceInfoHour {
	return &model.DeviceInfoHour{
		UserID:            device.UserID,
		RetrievalCount:    device.RetrievalCount,
		BlockCount:        device.CacheCount,
		DeviceID:          device.DeviceID,
		Time:              t,
		DiskUsage:         device.DiskUsage,
		DiskSpace:         device.DiskSpace,
		HourIncome:        device.CumulativeProfit,
		BandwidthUp:       device.BandwidthUp,
		BandwidthDown:     device.BandwidthDown,
		UpstreamTraffic:   device.UploadTraffic,
		DownstreamTraffic: device.DownloadTraffic,
		OnlineTime:        device.OnlineTime,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func ToDeviceInfo(node types.NodeInfo, areaId string) *model.DeviceInfo {
	deviceInfo := model.DeviceInfo{
		DeviceID:         node.NodeID,
		DeviceName:       node.NodeName,
		DiskSpace:        formatter.ToFixed(node.DiskSpace, 2),
		DiskUsage:        formatter.ToFixed(node.DiskUsage, 2),
		TitanDiskSpace:   formatter.ToFixed(node.AvailableDiskSpace, 2),
		TitanDiskUsage:   formatter.ToFixed(node.TitanDiskUsage, 2),
		ActiveStatus:     1,
		OnlineTime:       float64(node.OnlineDuration),
		BandwidthUp:      float64(node.BandwidthUp),
		BandwidthDown:    float64(node.BandwidthDown),
		CpuUsage:         formatter.ToFixed(node.CPUUsage, 2),
		CpuCores:         int64(node.CPUCores),
		Memory:           node.Memory,
		MemoryUsage:      formatter.ToFixed(node.MemoryUsage, 2),
		UploadTraffic:    float64(node.UploadTraffic),
		DownloadTraffic:  float64(node.DownloadTraffic),
		ExternalIp:       node.ExternalIP,
		InternalIp:       node.InternalIP,
		IoSystem:         node.IoSystem,
		SystemVersion:    node.SystemVersion,
		DiskType:         node.DiskType,
		CumulativeProfit: node.Profit,
		NodeType:         int64(node.Type),
		CacheCount:       node.AssetCount,
		RetrievalCount:   node.RetrieveCount,
		NATType:          node.NATType,
		UpdatedAt:        node.LastSeen,
		BoundAt:          node.FirstTime,
		IncomeIncr:       node.IncomeIncr,
		AreaID:           areaId,
	}

	switch node.Status {
	case types.NodeOffline:
		deviceInfo.DeviceStatus = DeviceStatusOffline
		deviceInfo.DeviceID = node.NodeID
		deviceInfo.DeviceStatusCode = DeviceStatusCodeOffline
	case types.NodeServicing, types.NodeNatSymmetric:
		deviceInfo.DeviceStatusCode = DeviceStatusCodeOnline
		deviceInfo.DeviceStatus = DeviceStatusOnline
	default:
		deviceInfo.DeviceStatusCode = DeviceStatusCodeAbnormal
		deviceInfo.DeviceStatus = DeviceStatusAbnormal
	}

	applyLocationInfo(&deviceInfo)

	return &deviceInfo
}

// applyLocationInfo 获取节点的 ip 位置信息
func applyLocationInfo(deviceInfo *model.DeviceInfo) {
	if deviceInfo.ExternalIp == "" {
		return
	}

	//var loc model.Location
	//err := GetIpLocation(context.Background(), deviceInfo.ExternalIp, &loc, model.LanguageCN, model.LanguageEN)
	loc, err := geo.GetIpLocation(context.Background(), deviceInfo.ExternalIp, model.LanguageEN)
	if err != nil || loc == nil {
		log.Errorf("get ip location %v", err)
		// applyLocationFromLocalGEODB(deviceInfo)
		return
	}

	deviceInfo.NetworkInfo = loc.Isp
	deviceInfo.IpProvince = loc.Province
	deviceInfo.IpCountry = loc.Country
	deviceInfo.IpCity = loc.City
	deviceInfo.IpLocation = dao.ContactIPLocation(*loc, model.LanguageEN)
	deviceInfo.Longitude, _ = strconv.ParseFloat(loc.Longitude, 64)
	deviceInfo.Latitude, _ = strconv.ParseFloat(loc.Latitude, 64)
}

var _ Fetcher = &NodeFetcher{}
