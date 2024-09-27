package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/node/cidutil"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/geo"
	"github.com/gnasnik/titan-explorer/core/oprds"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/core/storage"
	"github.com/shopspring/decimal"
)

const (
	maxTotalFlow    = 1 * 1024 * 1024 * 1024
	maxVipTotalFlow = 1000 * 1024 * 1024 * 1024
)

var (
	maxCountOfVisitAsset     int64 = 10
	maxCountOfVisitShareLink int64 = 10
	// AreaIDIPMaps 调度器区域id和ip映射
	AreaIDIPMaps = new(sync.Map)
	// AreaIPIDMaps 调度器区域ip和id映射
	AreaIPIDMaps = new(sync.Map)
	// CityAreaIDMaps 国家地区映射
	cityAreaIDMaps = make(map[string][]string)
	// lastSyncTimeStamp 上次同步时间
	lastSyncTimeStamp time.Time
	syncTimeMu        = new(sync.Mutex)
	rwMapsMu          = new(sync.RWMutex)
)

type (
	AssetRecord struct {
		CID                   string    `db:"cid"`
		Hash                  string    `db:"hash"`
		NeedEdgeReplica       int64     `db:"edge_replicas"`
		TotalSize             int64     `db:"total_size"`
		TotalBlocks           int64     `db:"total_blocks"`
		Expiration            time.Time `db:"expiration"`
		CreatedTime           time.Time `db:"created_time"`
		EndTime               time.Time `db:"end_time"`
		NeedCandidateReplicas int64     `db:"candidate_replicas"`
		ServerID              string    `db:"scheduler_sid"`
		State                 string    `db:"state"`
		NeedBandwidth         int64     `db:"bandwidth"` // unit:MiB/
		Note                  string    `db:"note"`
		Source                int64     `db:"source"`

		RetryCount        int64 `db:"retry_count"`
		ReplenishReplicas int64 `db:"replenish_replicas"`
		ReplicaNums       int64 `json:"replica_num"`

		SPCount int64
	}
	// AssetOverview 文件概览
	AssetOverview struct {
		AssetRecord      *AssetRecord
		UserAssetDetail  *dao.UserAssetDetail
		VisitCount       int64
		RemainVisitCount int64
	}
	// ListAssetRecordRsp list asset records
	ListAssetRecordRsp struct {
		Total          int64            `json:"total"`
		AssetOverviews []*AssetOverview `json:"asset_infos"`
	}

	// UserAssetSummary user asset and group
	UserAssetSummary struct {
		AssetOverview *AssetOverview
		AssetGroup    *dao.AssetGroup
	}
	// ListAssetSummaryRsp list asset and group
	ListAssetSummaryRsp struct {
		Total int64               `json:"total"`
		List  []*UserAssetSummary `json:"list"`
	}
)

func getAreaIDsByAreaID(c *gin.Context, areaIDs []string) ([]string, map[string][]string) {
	var (
		aids, naids []string
	)
	// 兼容以前的区域请求
	areaIDs = getAreaIDsCountry(areaIDs)

	_, maps, err := GetAndStoreAreaIDs()
	if err != nil {
		log.Error(err)
	}

	for _, v := range areaIDs {
		if strings.TrimSpace(v) != "" {
			aids = append(aids, maps[v]...)
		}
	}
	if len(aids) == 0 {
		for _, v := range maps {
			aids = append(aids, v...)
		}
	}

	if len(aids) == 1 {
		return aids, maps
	}

	sort.Slice(aids, func(i, j int) bool {
		return aids[i] < aids[j]
	})

	// 获取用户的访问的ip
	ip, err := GetIPFromRequest(c.Request)
	if err != nil {
		log.Errorf("get user's ip of request error:%w", err)
	} else {
		tadis := aids
		// 获取区域里的调度器
		info, err := geo.GetIpLocation(c.Request.Context(), ip)
		if err == nil {
			for _, v := range areaIDs {
				if strings.EqualFold(v, info.Country) {
					if vv, ok := maps[v]; ok {
						tadis = vv
					}
					break
				}
			}
		}
		areaID, err := GetNearestAreaID(c.Request.Context(), ip, tadis)
		if err != nil {
			log.Error(err)
		} else {
			// areaID = "Asia-China-Guangdong-Shenzhen"
			naids = append(naids, areaID)
			for _, v := range aids {
				if !strings.EqualFold(v, areaID) {
					naids = append(naids, v)
				}
			}
			return naids, maps
		}
	}

	return aids, maps
}

func getAreaIDs(c *gin.Context) []string {
	areaIDs := c.QueryArray("area_id")
	aids, _ := getAreaIDsByAreaID(c, areaIDs)

	return aids
}

func getAreaIDsByArea(c *gin.Context, areaIds []string) []string {
	aids, _ := getAreaIDsByAreaID(c, areaIds)

	return aids
}

func getAreaIDsNoDefault(c *gin.Context) []string {
	var aids []string

	_, maps, err := GetAndStoreAreaIDs()
	if err != nil {
		log.Error(err)
		return nil
	}

	areaIDs := getAreaIDsCountry(c.QueryArray("area_id"))
	for _, v := range areaIDs {
		v := strings.TrimSpace(v)
		if v != "" {
			aids = append(aids, maps[v]...)
		}
	}

	return aids
}

func getAreaID(c *gin.Context) string {
	areaID := strings.TrimSpace(c.Query("area_id"))
	areaID = GetDefaultTitanCandidateEntrypointInfo()

	if areaID == "" {
		areaID = GetDefaultTitanCandidateEntrypointInfo()
	} else {
		areaIds := getAreaIDsCountry([]string{c.Query("area_id")})
		aids, _ := getAreaIDsByAreaID(c, areaIds)
		if len(aids) > 0 {
			areaID = aids[0]
		} else {
			areaID = GetDefaultTitanCandidateEntrypointInfo()
		}
	}

	return areaID
}

func listAssets(ctx context.Context, uid string, limit, offset, groupID int) (*ListAssetRecordRsp, error) {
	var (
		wg = new(sync.WaitGroup)
		mu = new(sync.Mutex)
	)
	uInfo, err := dao.GetUserByUsername(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user's info error:%w", err)
	}
	total, infos, err := dao.ListAssets(ctx, uid, limit, offset, groupID)
	if err != nil {
		return nil, fmt.Errorf("get list of asset error:%w", err)
	}

	list := make([]*AssetOverview, len(infos))

	for i, info := range infos {
		wg.Add(1)
		go func(i int, info *dao.UserAssetDetail) {
			defer wg.Done()

			// 获取用户文件所有调度器区域
			areaIDs, err := dao.GetUserAssetAreaIDs(ctx, info.Hash, uid)
			if err != nil {
				log.Errorf("get areaids err: %s", err.Error())
				return
			}
			// 获取用户文件分发记录
			records := new(AssetRecord)
			cid := info.Cid
			if cid == "" {
				// 将 hash 转换为 cid
				cid, err = storage.HashToCID(info.Hash)
				if err != nil {
					log.Errorf("hash to cid err: %s", err.Error())
					return
				}
			} else {
				records.CID = cid
			}
			records.Hash = info.Hash
			for _, v := range areaIDs {
				sCli, err := getSchedulerClient(ctx, v)
				if err != nil {
					log.Errorf("getSchedulerClient err: %s", err.Error())
					continue
				}
				record, err := sCli.GetAssetRecord(ctx, cid)
				if err != nil {
					log.Errorf("asset LoadAssetRecord err: %s", err.Error())
					continue
				}
				if info.Cid == "" {
					dao.UpdateAssetCid(ctx, info.Hash, record.CID)
				}
				records.NeedEdgeReplica += record.NeedEdgeReplica
				records.NeedCandidateReplicas += record.ReplenishReplicas
				// records.ReplicaInfos = append(records.ReplicaInfos, record.ReplicaInfos...)
				records.ReplicaNums += int64(len(record.ReplicaInfos))
				// for _, vv := range record.ReplicaInfos {
				// 	if vv.Status == 3 {
				// 		records.ReplicaNums++
				// 	}
				// }
				if records.TotalSize == 0 {
					records.CID = record.CID
					records.CreatedTime = record.CreatedTime
					records.EndTime = record.EndTime
					records.Expiration = record.Expiration
					records.Note = record.Note
					records.ServerID = fmt.Sprintf("%v", record.ServerID)
					records.State = record.State
					records.Source = record.Source
					records.TotalBlocks = record.TotalBlocks
					records.TotalSize = record.TotalSize
				}
			}
			if !uInfo.EnableVIP && info.VisitCount >= maxCountOfVisitAsset {
				info.ShareStatus = 2
			}
			info.AreaIDs = append(info.AreaIDs, areaIDs...)
			r := &AssetOverview{
				AssetRecord:      records,
				UserAssetDetail:  info,
				VisitCount:       info.VisitCount,
				RemainVisitCount: maxCountOfVisitAsset - info.VisitCount,
			}
			mu.Lock()
			list[i] = r
			mu.Unlock()
		}(i, info)
	}
	wg.Wait()

	return &ListAssetRecordRsp{Total: total, AssetOverviews: list}, nil
}

// getAssetOverView Get only on asset view, copy from listAssets
func getAssetOverView(ctx context.Context, uid, cid string) (*AssetOverview, error) {
	uInfo, err := dao.GetUserByUsername(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user's info error:%w", err)
	}
	hash, err := cidutil.CIDToHash(cid)
	if err != nil {
		return nil, err
	}

	assetDetail, err := dao.GetUserAssetDetail(ctx, hash, uid)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error occurred while fetching asset details: %s", err)

	}
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("asset details not found : %s", err)
	}

	areaIDs, err := dao.GetUserAssetAreaIDs(ctx, assetDetail.Hash, uid)
	if err != nil {
		return nil, fmt.Errorf("get areaids err: %s", err.Error())
	}
	// 获取用户文件分发记录
	records := new(AssetRecord)
	if cid == "" {
		// 将 hash 转换为 cid
		cid, err = storage.HashToCID(assetDetail.Hash)
		if err != nil {
			return nil, fmt.Errorf("hash to cid err: %s", err.Error())
		}
	} else {
		records.CID = cid
	}

	records.Hash = assetDetail.Hash
	for _, v := range areaIDs {
		sCli, err := getSchedulerClient(ctx, v)
		if err != nil {
			log.Errorf("getSchedulerClient err: %s", err.Error())
			continue
		}
		record, err := sCli.GetAssetRecord(ctx, cid)
		if err != nil {
			log.Errorf("asset LoadAssetRecord err: %s", err.Error())
			continue
		}
		if assetDetail.Cid == "" {
			dao.UpdateAssetCid(ctx, assetDetail.Hash, record.CID)
		}
		records.NeedEdgeReplica += record.NeedEdgeReplica
		records.NeedCandidateReplicas += record.ReplenishReplicas
		// records.ReplicaInfos = append(records.ReplicaInfos, record.ReplicaInfos...)
		for _, vv := range record.ReplicaInfos {
			if vv.Status == 3 {
				records.ReplicaNums++
			}
		}
		if records.TotalSize == 0 {
			records.CID = record.CID
			records.CreatedTime = record.CreatedTime
			records.EndTime = record.EndTime
			records.Expiration = record.Expiration
			records.Note = record.Note
			records.ServerID = fmt.Sprintf("%v", record.ServerID)
			records.State = record.State
			records.Source = record.Source
			records.TotalBlocks = record.TotalBlocks
			records.TotalSize = record.TotalSize
		}
	}
	if !uInfo.EnableVIP && assetDetail.VisitCount >= maxCountOfVisitAsset {
		assetDetail.ShareStatus = 2
	}
	assetDetail.AreaIDs = append(assetDetail.AreaIDs, areaIDs...)
	r := &AssetOverview{
		AssetRecord:      records,
		UserAssetDetail:  assetDetail,
		VisitCount:       assetDetail.VisitCount,
		RemainVisitCount: maxCountOfVisitAsset - assetDetail.VisitCount,
	}

	return r, nil
}

func getAssetStatus(ctx context.Context, uid, cid string) (*types.AssetStatus, error) {
	var (
		resp       = new(types.AssetStatus)
		visitCount int64
	)

	// 获取用户信息
	uInfo, err := dao.GetUserByUsername(ctx, uid)
	switch err {
	case sql.ErrNoRows:
		uInfo = new(model.User)
	case nil:
	default:
		return nil, fmt.Errorf("get user's info error:%w", err)
	}
	// 获取链接信息
	linkInfo, err := dao.GetLink(ctx, squirrel.Select("*").Where("username = ?", uid).Where("cid = ?", cid))
	if err != nil {
		return nil, fmt.Errorf("get link info error:%w", err)
	}
	// 判断是文件组id还是文件id
	gid, _ := strconv.Atoi(cid)
	if gid == 0 {
		// 将cid转换为hash
		hash, err := storage.CIDToHash(cid)
		if err != nil {
			return nil, err
		}
		aInfo, err := dao.GetUserAssetDetail(ctx, hash, uid)
		switch err {
		case sql.ErrNoRows:
			aInfo = new(dao.UserAssetDetail)
		case nil:
			resp.IsExist = true
		default:
			return nil, fmt.Errorf("get asset's info error:%w", err)
		}
		resp.IsExpiration = aInfo.Expiration.Before(time.Now()) && linkInfo.ExpireAt.Before(time.Now())
		visitCount = aInfo.VisitCount
	} else {
		ginfo, err := dao.GetUserAssetGroupInfo(ctx, uid, gid)
		switch err {
		case sql.ErrNoRows:
		case nil:
			resp.IsExist = true
		default:
			return nil, fmt.Errorf("get asset's group info error:%w", err)
		}
		resp.IsExpiration = linkInfo.ExpireAt.Before(time.Now())
		visitCount = ginfo.VistitCount
	}

	if uInfo.EnableVIP {
		return resp, nil
	}
	if visitCount >= maxCountOfVisitShareLink {
		resp.IsVisitOutOfLimit = true
	}

	return resp, nil
}

func listAssetSummary(ctx context.Context, uid string, parent, page, size int) (*ListAssetSummaryRsp, error) {
	resp := new(ListAssetSummaryRsp)
	offset := (page - 1) * size
	groupRsp, err := dao.ListAssetGroupForUser(ctx, uid, parent, size, offset)
	if err != nil {
		return nil, err
	}

	for _, group := range groupRsp.AssetGroups {
		i := new(UserAssetSummary)
		i.AssetGroup = group
		resp.List = append(resp.List, i)
	}
	resp.Total = groupRsp.Total
	aLimit := size - len(groupRsp.AssetGroups)
	if aLimit < 0 {
		aLimit = 0
	}
	aOffset := offset - int(groupRsp.Total)
	if aOffset < 0 {
		aOffset = 0
	}

	assetRsp, err := listAssets(ctx, uid, aLimit, aOffset, parent)
	if err != nil {
		return nil, err
	}
	for _, asset := range assetRsp.AssetOverviews {
		i := new(UserAssetSummary)
		i.AssetOverview = asset
		resp.List = append(resp.List, i)
	}
	resp.Total += assetRsp.Total

	return resp, nil
}

// SyncAreaIDs 同步未登陆用户文件的区域
func SyncAreaIDs(ctx context.Context, sCli api.Scheduler, nodeID, cid string, size int64, areaIds []string) ([]string, error) {
	zStrs := make([]string, 0)
	if len(areaIds) == 0 {
		return zStrs, nil
	}

	info, err := sCli.GenerateTokenForDownloadSource(ctx, nodeID, cid)
	if err != nil {
		log.Errorf("generate token for download source error:%w", err)
		return zStrs, nil
	}
	for _, v := range areaIds {
		var repCount int64 = 5
		if len(areaIds) == 1 {
			repCount = 10
		}
		scli, err := getSchedulerClient(ctx, v)
		if err != nil {
			log.Errorf("getSchedulerClient error: %v", err)
			continue
		}
		err = scli.CreateSyncAsset(ctx, &types.CreateSyncAssetReq{
			AssetCID:      cid,
			AssetSize:     size,
			DownloadInfo:  info,
			ReplicaCount:  repCount,
			ExpirationDay: 1,
		})
		if err != nil {
			log.Errorf("GetUserAssetByAreaIDs error: %v", err)
			continue
		}
		zStrs = append(zStrs, v)
	}

	return zStrs, nil
}

// GetAreaIPByID 根据areaid信息获取调度器的ip
func GetAreaIPByID(ctx context.Context, areaID string) (string, error) {
	ip, ok := AreaIDIPMaps.Load(areaID)
	if ok {
		return ip.(string), nil
	}

	schedulers, err := statistics.GetSchedulerConfigs(ctx, fmt.Sprintf("%s::%s", SchedulerConfigKeyPrefix, areaID))
	if err != nil {
		return "", err
	}
	uri := schedulers[0].SchedulerURL
	aurl, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	uri, _, _ = strings.Cut(aurl.Host, ":")
	ips, err := net.LookupIP(uri)
	if err != nil {
		return "", nil
	}
	AreaIDIPMaps.Store(areaID, ips[0].String())
	AreaIPIDMaps.Store(ips[0].String(), areaID)

	return ips[0].String(), nil
}

// GetIPFromRequest 根据请求获取ip地址
func GetIPFromRequest(r *http.Request) (ip string, err error) {
	// 判断是否为内网IP或环回地址
	defer func() {
		ipv4 := net.ParseIP(ip)
		if ipv4.IsPrivate() || ipv4.IsLoopback() {
			ip = ""
			err = errors.New("internal ip or loop back ip")
		}
	}()

	// 检查 X-Forwarded-For 头
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个IP地址，取第一个
		ips := strings.Split(ip, ",")
		clientIP := strings.TrimSpace(ips[0])
		if net.ParseIP(clientIP) != nil {
			return clientIP, nil
		}
	}

	// 检查 X-Real-IP 头
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	// 如果没有代理服务器，则使用 RemoteAddr
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("userip: %q is not IP:port,error:%w", r.RemoteAddr, err)
	}

	return ip, nil
}

// GetNearestAreaIDByIP 根据ip获取距离用户最近的areaid
func GetNearestAreaIDByIP(ctx context.Context, ip string, areaIDs []string) (string, error) {
	var ips []string

	// 将areaid替换为ip
	for _, v := range areaIDs {
		ip, err := GetAreaIPByID(ctx, v)
		if err != nil {
			continue
		}
		ips = append(ips, ip)
	}
	log.Errorf("user ip:%s ips:%v", ip, ips)

	ip, err := GetUserFixedNearestIP(ctx, ip, ips, NewIPCoordinate())
	if err != nil {
		return "", err
	}
	log.Errorf("nearest ip:%s", ip)

	if areaID, ok := AreaIPIDMaps.Load(ip); ok {
		return areaID.(string), nil
	}

	return "", errors.New("not found")
}

// GetNearestAreaIDByInfo 根据ip的相关位置信息获取距离用户最近的areaid
func GetNearestAreaIDByInfo(ctx context.Context, ip string, areaIDs []string) (string, error) {
	var existAreaIDs []string
	info, err := geo.GetIpLocation(ctx, ip)
	if err != nil {
		return "", fmt.Errorf("get info of ip error:%w", err)
	}
	log.Error(info.Continent, info.Country)

	for _, v := range areaIDs {
		if strings.Contains(v, info.Continent) {
			existAreaIDs = append(existAreaIDs, v)
		}
	}
	if len(existAreaIDs) > 0 {
		for _, v := range existAreaIDs {
			if strings.Contains(v, info.Country) {
				return v, nil
			}
		}
	}

	return "", errors.New("not found")
}

// GetNearestAreaID 聚合获距离用户请求的最近的areaid
func GetNearestAreaID(ctx context.Context, ip string, areaIDs []string) (string, error) {
	areaID, err := GetNearestAreaIDByIP(ctx, ip, areaIDs)
	if err == nil {
		return areaID, nil
	}
	log.Errorf("get nearest areaid error:%v", err)

	return GetNearestAreaIDByInfo(ctx, ip, areaIDs)
}

// GetFILPrice 获取filecoin的价格
func GetFILPrice(ctx context.Context) (float64, error) {
	var priceMap = make(map[string]interface{})

	key := "FIL_price"
	// 先从redis缓存中获取，获取不到再请求url
	price, _ := oprds.GetClient().RedisClient().Get(ctx, key).Float64()
	if price != 0 {
		return price, nil
	}

	resp, err := http.Get("https://api.coincap.io/v2/assets/filecoin")
	if err != nil {
		return 0, fmt.Errorf("get price of filecoin error:%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body of response error:%w", err)
	}

	err = json.Unmarshal(body, &priceMap)
	if err != nil {
		return 0, fmt.Errorf("json unmarshal error:%w", err)
	}

	if dataMap, ok := priceMap["data"]; ok {
		if data, ok := dataMap.(map[string]interface{}); ok {
			if priceStr, ok := data["priceUsd"].(string); ok {
				dc, err := decimal.NewFromString(priceStr)
				if err != nil {
					return 0, fmt.Errorf("decimal error:%w", err)
				}
				price, _ := dc.Round(2).Float64()
				oprds.GetClient().RedisClient().SetEx(ctx, key, price, 5*time.Minute)
				return price, nil
			}
		}
	}

	return 0, errors.New("get price of filecoin error")
}

// GetAndStoreAreaIDs 获取或存储节点地区信息
func GetAndStoreAreaIDs() ([]string, map[string][]string, error) {
	tn := time.Now()
	syncTimeMu.Lock()
	defer syncTimeMu.Unlock()
	if lastSyncTimeStamp.IsZero() {
		lastSyncTimeStamp = time.Now()
	} else {
		lt := lastSyncTimeStamp.Add(10 * time.Minute)
		if lt.After(tn) {
			keys, maps := rangeCityAidMaps()
			return keys, maps, nil
		}
	}

	etcdClient, err := statistics.NewEtcdClient(config.Cfg.EtcdAddresses)
	if err != nil {
		keys, maps := rangeCityAidMaps()
		return keys, maps, fmt.Errorf("New etcdClient Failed: %w", err)
	}
	schedulers, err := statistics.FetchSchedulersFromEtcd(etcdClient)
	if err != nil {
		keys, maps := rangeCityAidMaps()
		return keys, maps, fmt.Errorf("fetch scheduler from etcd Failed: %w", err)
	}
	rwMapsMu.Lock()
	cityAreaIDMaps = make(map[string][]string)
	for _, v := range schedulers {
		as := strings.Split(v.AreaId, "-")
		if len(as) < 2 {
			continue
		}
		cityAreaIDMaps[as[1]] = append(cityAreaIDMaps[as[1]], v.AreaId)
	}
	rwMapsMu.Unlock()

	keys, maps := rangeCityAidMaps()
	return keys, maps, nil
}

func rangeCityAidMaps() ([]string, map[string][]string) {
	var (
		keys []string
		maps = make(map[string][]string)
	)

	rwMapsMu.RLock()
	for k, v := range cityAreaIDMaps {
		keys = append(keys, k)
		maps[k] = append(maps[k], v...) // 拷贝每个 key 的 slice
	}
	rwMapsMu.RUnlock()

	return keys, maps
}

func operateAreaIDs(ctx context.Context, areaIDs []string) []string {
	var aids []string

	for _, v := range areaIDs {
		as := strings.Split(v, "-")
		if len(as) < 2 {
			aids = append(aids, v)
		} else {
			aids = append(aids, as[1])
		}
	}

	return aids
}

func operateAreaMaps(ctx context.Context, aids []string, lan string) []dao.KVMap {
	var kvs = make([]dao.KVMap, 0)

	if lan == "cn" {
		maps, _ := dao.GetAreaMapByEn(ctx, aids)
		for _, v := range maps {
			kvs = append(kvs, dao.KVMap{
				Key:   v.AreaCn,
				Value: v.AreaEn,
			})
		}
	} else {
		for _, v := range aids {
			kvs = append(kvs, dao.KVMap{
				Key:   v,
				Value: v,
			})
		}
	}

	return kvs
}

// getAreaIDsCountry 兼容以前的区域请求，获取区域的country
func getAreaIDsCountry(areaIDs []string) []string {
	var (
		newAreaIDs []string
		areaMaps   = make(map[string]bool)
	)

	for _, v := range areaIDs {
		v = strings.TrimSpace(v)
		vs := strings.Split(v, "-")
		vv := v
		if len(vs) >= 2 {
			vv = vs[1]
		}
		areaMaps[vv] = false
	}
	for k := range areaMaps {
		newAreaIDs = append(newAreaIDs, k)
	}

	return newAreaIDs
}

// checkUserTotalFlow 判断用户使用总流量是否到达最大限制
func checkUserTotalFlow(ctx context.Context, username string) (bool, error) {
	var (
		fInfo = new(dao.UserStorageFlowInfo)
	)

	// 获取用户信息，判断用户是否为vip
	user, err := dao.GetUserByUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("get userInfo error:%w", err)
	}

	// 获取用户已使用的总流量
	value, err := oprds.GetClient().GetUserStorageFlowInfo(ctx, username)
	if err != nil {
		fInfo, err = dao.GetUserStorageFlowInfo(ctx, username)
		if err != nil {
			fInfo = new(dao.UserStorageFlowInfo)
			return false, err
		}
		ib, _ := json.Marshal(fInfo)
		oprds.GetClient().StoreUserStorageFlowInfo(ctx, username, string(ib))
	} else {
		json.Unmarshal([]byte(value), fInfo)
	}
	// 判断是否为vip
	switch user.EnableVIP {
	case true:
		if fInfo.TotalTraffic < maxVipTotalFlow {
			return true, nil
		}
	default:
		if fInfo.TotalTraffic < maxTotalFlow {
			return true, nil
		}
	}

	return false, nil
}

// 判断 apikey 是否存在
func checkAPIKeyIsExist(apiKey, uid string) (bool, error) {
	info, err := dao.GetUserByUsername(context.Background(), uid)
	if err != nil {
		return false, fmt.Errorf("get user's info error:%w", err)
	}

	if len(info.ApiKeys) <= 0 {
		return false, nil
	}

	keyResp, err := storage.DecodeAPIKeySecrets(info.ApiKeys)
	if err != nil {
		return false, fmt.Errorf("decode api secrets error:%w", err)
	}
	for _, v := range keyResp {
		if v.APIKey == apiKey {
			return true, nil
		}
	}

	return false, nil
}

func checkAuthGetGroup(ctx context.Context, uid string, gid int) error {
	uInfo, err := dao.GetUserByUsername(ctx, uid)
	linkInfo, err := dao.GetLink(ctx, squirrel.Select("*").Where("username = ?", uid).Where("cid = ?", gid))
	if err != nil {
		return fmt.Errorf("api GetLink: %w", err)
	}
	ginfo, err := dao.GetUserAssetGroupInfo(ctx, uid, gid)
	if err != nil {
		return fmt.Errorf("api GetUserAssetGroupInfo: %w", err)
	}
	if linkInfo.ExpireAt.Before(time.Now()) {
		return errors.New("expire")
	}

	if !uInfo.EnableVIP && ginfo.VistitCount >= maxCountOfVisitShareLink {
		return errors.New("out limit of share")
	}

	return nil
}
