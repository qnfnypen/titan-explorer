package job

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/oprds"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	logging "github.com/ipfs/go-log/v2"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"golang.org/x/exp/rand"
)

var (
	ctx         = context.Background()
	l1States    = []string{"EdgesSelect", "EdgesPulling", "Servicing", "EdgesFailed"}
	cronLog     = logging.Logger("cron")
	redisClient *goredislib.Client
	once        sync.Once
)

// 初始化 Redis 客户端
func getRedisCli() *goredislib.Client {
	once.Do(func() {
		redisClient = goredislib.NewClient(&goredislib.Options{
			Addr:     config.Cfg.RedisAddr,
			Password: config.Cfg.RedisPassword,
		})
	})
	return redisClient
}

func newRedSync() *redsync.Redsync {
	pool := goredis.NewPool(getRedisCli())
	return redsync.New(pool)
}

// SyncShedulersAsset 同步调度器文件
func SyncShedulersAsset() {
	c := cron.New(cron.WithLocation(time.Local))

	// 初始化分布式锁
	redsync := newRedSync()

	c.AddFunc("@every 10s", func() {
		// 防止同时竞争一把锁
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		mutex := redsync.NewMutex("syncUserScheduler-lock")
		if err := mutex.Lock(); err != nil {
			log.Printf("syncUserScheduler is already running on another instance: %v", err)
			return
		}
		syncUserScheduler()
	})
	c.AddFunc("@every 15s", func() {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		mutex := redsync.NewMutex("syncUnLoginAsset-lock")
		if err := mutex.Lock(); err != nil {
			log.Printf("syncUnLoginAsset is already running on another instance: %v", err)
			return
		}
		syncUnLoginAsset()
	})
	c.AddFunc("0,10,20,30,40,50 * * * *", func() {
		// 防止同时竞争一把锁
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		mutex := redsync.NewMutex("syncDashboard-lock")
		if err := mutex.Lock(); err != nil {
			log.Printf("syncDashboard is already running on another instance: %v", err)
			return
		}
		syncDashboard()
	})
	c.AddFunc("@every 60s", func() {
		// 防止同时竞争一把锁
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		mutex := redsync.NewMutex("getSyncSuccessAsset-lock")
		if err := mutex.Lock(); err != nil {
			log.Printf("getSyncSuccessAsset is already running on another instance: %v", err)
			return
		}
		getSyncSuccessAsset()
	})

	c.Start()
}

// syncUserScheduler 同步登陆后用户的调度器信息
func syncUserScheduler() {
	// 获取 schedulers
	payloads, err := oprds.GetClient().GetAllSchedulerInfos(ctx)
	if err != nil {
		cronLog.Errorf("get all scheduler infos error:%v", err)
		return
	}
	wg := new(sync.WaitGroup)
	for _, v := range payloads {
		wg.Add(1)
		go func(v *oprds.Payload) {
			defer wg.Done()

			scli, err := api.GetSchedulerClient(ctx, v.AreaID)
			if err != nil {
				cronLog.Errorf("get client of scheduler error:%v", err)
				return
			}
			// 判断L1节点是否同步完成
			rs, err := scli.GetAssetRecord(ctx, v.CID)
			if err != nil {
				cronLog.Errorf("GetAssetRecord error:%v", err)
				return
			}
			if !checkSyncState(rs.State) {
				return
			}
			unSyncAids, err := dao.GetUnSyncAreaIDs(ctx, v.UserID, v.Hash)
			if err != nil {
				cronLog.Errorf("GetUnSyncAreaIDs error:%v", err)
				return
			}
			// 对于已经有的节点不再进行同步，并变更状态
			aids, err := SyncShedulers(ctx, scli, v.CID, 0, v.Owner, unSyncAids)
			if err != nil {
				cronLog.Errorf("SyncShedulers error:%v", err)
				return
			}
			err = dao.UpdateUnSyncAreaIDs(ctx, v.UserID, v.Hash, aids)
			if err != nil {
				log.Println(fmt.Errorf("UpdateUnSyncAreaIDs error:%w", err))
				return
			}
			oprds.GetClient().DelSchedulerInfo(ctx, v)
		}(v)
	}
	wg.Wait()
}

func syncUnLoginAsset() {
	// 获取 schedulers
	payloads, err := oprds.GetClient().GetAllAreaIDs(ctx)
	if err != nil {
		cronLog.Errorf("get all scheduler infos error:%v", err)
		return
	}
	wg := new(sync.WaitGroup)
	for _, v := range payloads {
		wg.Add(1)
		go func(v *oprds.AreaIDPayload) {
			defer wg.Done()

			scli, err := api.GetSchedulerClient(ctx, v.AreaIDs[0])
			if err != nil {
				cronLog.Errorf("get client of scheduler error:%v", err)
				return
			}
			// 判断L1节点是否同步完成
			rs, err := scli.GetAssetRecord(ctx, v.CID)
			if err != nil {
				cronLog.Errorf("GetAssetRecord error:%v", err)
				return
			}
			if len(rs.ReplicaInfos) == 0 {
				return
			}
			aids, err := api.SyncAreaIDs(ctx, scli, "", v.CID, 0, v.AreaIDs[1:])
			if err != nil {
				cronLog.Errorf("SyncShedulers error:%v", err)
				return
			}
			aids = append(aids, v.AreaIDs[0])
			payload := oprds.UnLoginSyncArea{}
			for _, v := range aids {
				payload.List = append(payload.List, oprds.UnloginSyncAreaDetail{AreaID: v, IsSync: true})
			}
			oprds.GetClient().SetUnloginAssetInfo(ctx, v.Hash, &payload)
			oprds.GetClient().DelAreaIDs(ctx, v)
		}(v)
	}
	wg.Wait()
}

func syncDashboard() {
	var (
		wg            = new(sync.WaitGroup)
		trafficMaps   = new(sync.Map)
		bandwidthMaps = new(sync.Map)
	)

	// 获取当前时间
	now := time.Now()
	pendTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	startTime := pendTime.Add(-10 * time.Minute)
	endTime := pendTime.Add(-1 * time.Second)

	areaIDs, err := getAllAreaIDs()
	if err != nil {
		cronLog.Errorf("get all areaids error:%v", err)
		return
	}

	for _, v := range areaIDs {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()

			scli, err := api.GetSchedulerClient(ctx, v)
			if err != nil {
				cronLog.Errorf("get client of scheduler error:%v", err)
				return
			}
			infos, err := scli.GetDownloadResultsFromAssets(ctx, nil, startTime, endTime)
			if err != nil {
				cronLog.Errorf("GetDownloadResultsFromAssets error:%v", err)
				return
			}
			// 取出每个hash的最大值
			for _, v := range infos {
				key := fmt.Sprintf("%s_%s", v.UserID, v.Hash)
				storeTfOrBw(trafficMaps, key, v.TotalTraffic)
				storeTfOrBw(bandwidthMaps, key, v.PeakBandwidth)
			}
		}(v)
	}
	wg.Wait()

	err = storeAssetHourStorages(trafficMaps, bandwidthMaps, pendTime)
	if err != nil {
		cronLog.Errorf("storeAssetHourStorages error:%v", err)
	}
}

// getSyncSuccessAsset 更新
func getSyncSuccessAsset() {
	var (
		wg = new(sync.WaitGroup)
	)

	// 获取所有调度器区域
	areaIDs, err := getAllAreaIDs()
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range areaIDs {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()

			hashs, err := getSyncSuccessHash(v)
			if err != nil {
				cronLog.Errorf("getSyncSuccessHash error:%v", err)
				return
			}
			err = dao.UpdateSyncAssetAreas(ctx, v, hashs)
			if err != nil {
				cronLog.Errorf("UpdateSyncAssetAreas error:%v", err)
			}
		}(v)
	}
	wg.Wait()
}
