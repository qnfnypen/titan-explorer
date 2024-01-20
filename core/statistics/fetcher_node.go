package statistics

import (
	"context"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	"net"
	"strconv"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/oschwald/geoip2-golang"
)

const maxPageSize = 100

var supportLanguages = []model.Language{model.LanguageCN, model.LanguageEN}

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
func (n *NodeFetcher) Fetch(ctx context.Context, scheduler *Scheduler) error {
	log.Info("start fetching all nodes")
	start := time.Now()
	defer func() {
		log.Infof("fetch all nodes done, cost: %v", time.Since(start))
	}()

	var total int64
	page, size := 1, maxPageSize

loop:
	offset := (page - 1) * size
	resp, err := scheduler.Api.GetNodeList(ctx, offset, size)
	if err != nil {
		log.Errorf("api ListNodes: %v", err)
		return nil
	}

	total += int64(len(resp.Data))
	page++

	var nodes []*model.DeviceInfo
	for _, node := range resp.Data {
		if node.NodeID == "" {
			continue
		}

		nodeInfo := toDeviceInfo(ctx, node)
		if nodeInfo.DeviceStatus == DeviceStatusOffline {
			// just update device status
			err = dao.UpdateDeviceStatus(ctx, nodeInfo)
			if err != nil {
				log.Errorf("update device status: %v", err)
			}
		}
		nodes = append(nodes, nodeInfo)
	}

	if len(nodes) < 1 {
		log.Errorf("start to fetch all nodes: nodes length is 0")
		return nil
	}

	log.Infof("handling %d/%d nodes", total, resp.Total)

	n.Push(ctx, func() error {
		e := dao.BulkUpsertDeviceInfo(ctx, nodes)
		if e != nil {
			log.Errorf("bulk upsert device info: %v", e)
		}

		if e = addDeviceInfoHours(ctx, nodes); err != nil {
			log.Errorf("add device info hours: %v", err)
		}
		return nil
	})

	if total < resp.Total {
		goto loop
	}

	// add inactive node records for statistics
	e := dao.GenerateInactiveNodeRecords(context.Background(), start)
	if e != nil {
		log.Errorf("generate inactive node records: %v", e)
	}

	return nil
}

func toDeviceInfo(ctx context.Context, node types.NodeInfo) *model.DeviceInfo {
	deviceInfo := model.DeviceInfo{
		DeviceID:         node.NodeID,
		DeviceName:       node.NodeName,
		DiskSpace:        formatter.ToFixed(node.DiskSpace, 2),
		DiskUsage:        formatter.ToFixed(node.DiskUsage, 2),
		ActiveStatus:     1,
		OnlineTime:       float64(node.OnlineDuration),
		BandwidthUp:      float64(node.BandwidthUp),
		BandwidthDown:    float64(node.BandwidthDown),
		CpuUsage:         formatter.ToFixed(node.CPUUsage, 2),
		MemoryUsage:      formatter.ToFixed(node.MemoryUsage, 2),
		UploadTraffic:    float64(node.UploadTraffic),
		DownloadTraffic:  float64(node.DownloadTraffic),
		ExternalIp:       node.ExternalIP,
		InternalIp:       node.InternalIP,
		CumulativeProfit: node.Profit,
		NodeType:         int64(node.Type),
		CacheCount:       node.AssetCount,
		RetrievalCount:   node.RetrieveCount,
		BoundAt:          node.FirstTime,
	}

	switch node.Status {
	case 0:
		deviceInfo.DeviceStatus = DeviceStatusOffline
		deviceInfo.DeviceID = node.NodeID
		deviceInfo.DeviceStatusCode = DeviceStatusCodeOffline
	case 1:
		deviceInfo.DeviceStatusCode = DeviceStatusCodeOnline
		deviceInfo.DeviceStatus = DeviceStatusOnline
	default:
		deviceInfo.DeviceStatusCode = DeviceStatusCodeAbnormal
		deviceInfo.DeviceStatus = DeviceStatusAbnormal
	}

	applyLocationInfo(&deviceInfo)

	return &deviceInfo
}

func applyLocationInfo(deviceInfo *model.DeviceInfo) {
	if deviceInfo.ExternalIp == "" {
		return
	}

	var loc model.Location
	err := GetIpLocation(context.Background(), deviceInfo.ExternalIp, &loc, model.LanguageEN)
	if err != nil {
		log.Errorf("%v", err)
		applyLocationFromLocalGEODB(deviceInfo)
		return
	}

	deviceInfo.NetworkInfo = loc.Isp
	deviceInfo.IpProvince = loc.Province
	deviceInfo.IpCountry = loc.Country
	deviceInfo.IpCity = loc.City
	deviceInfo.IpLocation = dao.ContactIPLocation(loc, model.LanguageEN)

	//continent := loc.Continent
	//deviceInfo.IpLocation = continent + "-" + deviceInfo.IpCountry + "-" + deviceInfo.IpProvince
	//if deviceInfo.IpCity != "" {
	//	deviceInfo.IpLocation += "-" + deviceInfo.IpCity
	//}

	deviceInfo.Longitude, _ = strconv.ParseFloat(loc.Longitude, 64)
	deviceInfo.Latitude, _ = strconv.ParseFloat(loc.Latitude, 64)
}

func applyLocationFromLocalGEODB(deviceInfo *model.DeviceInfo) {
	db, err := geoip2.Open("city.mmdb")
	if err != nil {
		log.Errorf("open city.mmdb: %v", err)
		return
	}
	defer db.Close()

	// If you are using strings that may be invalid, check that ip is not nil
	if deviceInfo.ExternalIp == "" {
		return
	}

	ip := net.ParseIP(deviceInfo.ExternalIp)
	record, err := db.City(ip)
	if err != nil {
		log.Errorf("query ip %s: %v", deviceInfo.ExternalIp, err)
		return
	}

	if len(record.Subdivisions) > 0 {
		deviceInfo.IpProvince = record.Subdivisions[0].Names["en"]
	}

	continent := record.Continent.Names["en"]
	deviceInfo.IpCountry = record.Country.Names["en"]
	deviceInfo.IpCity = record.City.Names["en"]
	deviceInfo.IpLocation = continent + "-" + deviceInfo.IpCountry + "-" + deviceInfo.IpProvince
	if deviceInfo.IpCity != "" {
		deviceInfo.IpLocation += "-" + deviceInfo.IpCity
	}

	deviceInfo.Longitude = record.Location.Longitude
	deviceInfo.Latitude = record.Location.Latitude

	return
}

func GetIpLocation(ctx context.Context, ip string, Loc *model.Location, languages ...model.Language) error {
	// get info from databases
	err := dao.GetLocationInfoByIp(ctx, ip, Loc, model.LanguageEN)
	if err != nil {
		return err
	}
	if Loc.Ip != "" {
		return nil
	}

	if len(languages) == 0 {
		languages = supportLanguages
	}

	for _, l := range languages {
		loc, err := iptool.IPDataCloudGetLocation(ctx, config.Cfg.IpDataCloud.Url, ip, config.Cfg.IpDataCloud.Key, string(l))
		if err != nil {
			log.Errorf("iptablecloud get location: %v", err)
			continue
		}
		if err := dao.UpsertLocationInfo(ctx, loc, l); err != nil {
			continue
		}

		*Loc = *loc
	}

	return nil
}

var _ Fetcher = &NodeFetcher{}
