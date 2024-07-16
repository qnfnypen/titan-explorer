package statistics

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/bsm/redislock"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"golang.org/x/net/context"
)

var log = logging.Logger("statistics")

const LockerTTL = 30 * time.Second
const statisticLockerKeyPrefix = "TITAN::STATISTIC"

var SumDevicesInterval = time.Second * 5

// FetcherRegistry to keep track of registered fetchers
var FetcherRegistry []func() Fetcher

// RegisterFetcher allows registering new fetchers
func RegisterFetcher(fetcher func() Fetcher) {
	FetcherRegistry = append(FetcherRegistry, fetcher)
}

// Statistic represents the statistics manager.
type Statistic struct {
	ctx        context.Context
	cfg        config.StatisticsConfig
	cron       *cron.Cron
	locker     *redislock.Client
	fetchers   []Fetcher
	slk        sync.Mutex
	schedulers []*Scheduler
	etcdClient *EtcdClient
}

// New creates a new Statistic instance.
func New(cfg config.StatisticsConfig, client *EtcdClient) *Statistic {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(time.Local),
	)

	// 从 Etcd 上获取所有调度器的配置
	schedulers, err := FetchSchedulersFromEtcd(client)
	if err != nil {
		log.Fatalf("fetch scheduler from etcd Failed: %v", err)
	}

	s := &Statistic{
		ctx:        context.Background(),
		cron:       c,
		cfg:        cfg,
		schedulers: schedulers,
		locker:     redislock.New(dao.RedisCache),
		fetchers:   make([]Fetcher, 0),
		etcdClient: client,
	}

	//  监听 Etcd 的变化，更新调度器
	go s.watchEtcdSchedulerConfig()

	// 加载注册的数据拉取任务
	for _, fetcher := range FetcherRegistry {
		s.fetchers = append(s.fetchers, fetcher())
	}

	return s
}

func (s *Statistic) UpdateSchedulers(schedulers []*Scheduler) {
	s.slk.Lock()
	defer s.slk.Unlock()

	s.schedulers = schedulers
}

// Run starts the cron jobs for statistics.
func (s *Statistic) Run() {
	if s.cfg.Disable {
		return
	}

	s.cron.AddFunc(s.cfg.Crontab, s.Once("FETCHER", s.runFetchers))
	s.cron.Start()
	s.handleJobs()
}

// handleJobs Fetcher的任务队列，对于第个不同类型的数据下载,都开一个协程处理,使用队列是避免并行写入数据库,获取不到锁写入失败的问题
func (s *Statistic) handleJobs() {
	for _, fetcher := range s.fetchers {
		go func(f Fetcher) {
			for {
				select {
				case job := <-f.GetJobQueue():
					t := reflect.TypeOf(f)
					log.Infof("%v jobqueue count: %d", t, len(f.GetJobQueue()))
					if err := job(); err != nil {
						log.Errorf("run job: %v", err)
					}

					// 当执行完所有的任务之后,调用 Finalize, 主要用于拉取节点数据完成之后的整个节点数据重新统计
					if len(f.GetJobQueue()) == 0 && hasFinished(f) {
						err := f.Finalize()
						if err != nil {
							log.Errorf("handle finalize: %v", err)
						}
					}

				case <-s.ctx.Done():
					return
				}
			}
		}(fetcher)
	}
}

func hasFinished(f Fetcher) bool {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	confirmed := 0

	for {
		select {
		case <-ticker.C:
			if confirmed >= 10 {
				return true
			}

			if len(f.GetJobQueue()) > 0 {
				return false
			}

			confirmed++
		}
	}
}

// runFetchers 遍历所有的调度器,每个调度器开一个协程,并行处理任务
func (s *Statistic) runFetchers() error {
	var wg sync.WaitGroup
	wg.Add(len(s.schedulers))
	s.slk.Lock()
	for _, scheduler := range s.schedulers {
		go func(scheduler *Scheduler) {
			defer wg.Done()
			for _, fetcher := range s.fetchers {
				err := fetcher.Fetch(s.ctx, scheduler)
				if err != nil {
					log.Errorf("run fetcher: %v", err)
				}
			}
		}(scheduler)
	}
	s.slk.Unlock()
	wg.Wait()

	return nil
}

// Stop stops the cron jobs and closes schedulers.
func (s *Statistic) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	}
	for _, scheduler := range s.schedulers {
		scheduler.Closer()
	}
}

// Once 使用 redis 分布式锁, 当部署多个服务时,保证只有一个服务进行数据拉取和统计,避免重复执行任务,获得锁的服务会执行任务,获取不到锁的则跳过.
func (s *Statistic) Once(key string, fn func() error) func() {
	return func() {
		dKey := fmt.Sprintf("%s::%s", statisticLockerKeyPrefix, key)
		lock, err := s.locker.Obtain(s.ctx, dKey, LockerTTL, nil)
		if err == redislock.ErrNotObtained {
			log.Debugf("%s: %v", dKey, redislock.ErrNotObtained)
			return
		}

		if err != nil {
			log.Errorf("obtain redis lock: %v", err)
			return
		}

		defer lock.Release(s.ctx)
		if err = fn(); err != nil {
			log.Errorf("execute cron job: %v", err)
		}
	}
}

// watchEtcdSchedulerConfig 监听 Etcd 中调度器的增加和减少变化
func (s *Statistic) watchEtcdSchedulerConfig() {
	watchChan := s.etcdClient.cli.WatchServers(context.Background(), types.NodeScheduler.String())
	for {
		resp, ok := <-watchChan
		if !ok {
			log.Errorf("close watch chan")
			return
		}

		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.DELETE, mvccpb.PUT:
				log.Infof("Etcd scheduler config changed")
				schedulers, err := FetchSchedulersFromEtcd(s.etcdClient)
				if err != nil {
					log.Errorf("FetchSchedulersFromEtcd: %v", err)
					continue
				}

				s.UpdateSchedulers(schedulers)
				log.Infof("Updated scheduler from etcd")
			}
		}
	}
}
