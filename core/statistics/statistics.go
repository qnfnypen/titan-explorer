package statistics

import (
	"fmt"
	"github.com/bsm/redislock"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var log = logging.Logger("statistics")

const (
	megaBytes = 1 << 20
	gigaBytes = 1 << 30
	teraBytes = 1 << 40
)

const LockerTTL = 30 * time.Second
const DKeyRunFetchers = "titan::dk_run_fetchers"

type Statistic struct {
	ctx        context.Context
	cfg        config.StatisticsConfig
	cron       *cron.Cron
	locker     *redislock.Client
	fetchers   []Fetcher
	schedulers []*Scheduler
}

func New(cfg config.StatisticsConfig, scheduler []*Scheduler) *Statistic {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(time.Local),
	)

	s := &Statistic{
		ctx:        context.Background(),
		cron:       c,
		cfg:        cfg,
		schedulers: scheduler,
		locker:     redislock.New(dao.Cache),
		fetchers: []Fetcher{
			newNodeFetcher(),
			newCacheFetcher(),
			newValidationFetcher(),
			newSystemInfoFetcher(),
		},
	}

	return s
}

func (s *Statistic) Run() {
	if s.cfg.Disable {
		return
	}
	s.cron.AddFunc(s.cfg.Crontab, s.runFetchers)
	s.cron.Start()
	s.handleJobs()
}

func (s *Statistic) handleJobs() {
	for _, fetcher := range s.fetchers {
		go func(f Fetcher) {
			for {
				select {
				case job := <-f.GetJobQueue():
					if err := job(); err != nil {
						log.Errorf("run job: %v", err)
					}
				case <-s.ctx.Done():
					return
				}
			}
		}(fetcher)
	}
}

func (s *Statistic) runFetchers() {
	var wg sync.WaitGroup
	wg.Add(len(s.schedulers))
	for _, scheduler := range s.schedulers {
		go func(scheduler *Scheduler) {
			defer wg.Done()
			s.Once(fmt.Sprintf("%s::%s", DKeyRunFetchers, scheduler.Uuid), func() error {
				for _, fetcher := range s.fetchers {
					err := fetcher.Fetch(s.ctx, scheduler)
					if err != nil {
						log.Errorf("run fetcher: %v", err)
					}
				}
				return nil
			})()
		}(scheduler)
	}
	wg.Wait()

	s.asyncExecute(
		[]func() error{
			s.SumDeviceInfoProfit,
			s.CountRetrievals,
			s.SumFullNodeInfo,
		},
	)
}

func (s *Statistic) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	}
	for _, scheduler := range s.schedulers {
		scheduler.Closer()
	}
}

func (s *Statistic) Once(key string, fn func() error) func() {
	return func() {
		lock, err := s.locker.Obtain(s.ctx, key, LockerTTL, nil)
		if err == redislock.ErrNotObtained {
			log.Debug(redislock.ErrNotObtained)
			return
		}

		if err != nil {
			log.Fatalf("obtain redis lock: %v", err)
			return
		}

		defer lock.Release(s.ctx)
		if err = fn(); err != nil {
			log.Errorf("execute cron job: %v", err)
		}
	}
}

func (s *Statistic) asyncExecute(jobs []func() error) {
	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for _, job := range jobs {
		go func(job func() error) {
			defer wg.Done()
			if err := job(); err != nil {
				log.Errorf("handling job: %v", err)
			}
		}(job)
	}
	wg.Wait()
}

func (s *Statistic) SumFullNodeInfo() error {
	fullNodeInfo, err := dao.SumFullNodeInfoFromDeviceInfo(s.ctx)
	if err != nil {
		log.Errorf("count full node: %v", err)
		return err
	}

	systemInfo, err := dao.SumSystemInfo(s.ctx)
	if err != nil {
		log.Errorf("sum system info: %v", err)
		return err
	}

	fullNodeInfo.TotalCarfile = systemInfo.CarFileCount
	fullNodeInfo.RetrievalCount = systemInfo.DownloadCount
	fullNodeInfo.NextElectionTime = time.Unix(systemInfo.NextElectionTime, 0)

	fullNodeInfo.Time = time.Now()
	fullNodeInfo.CreatedAt = time.Now()
	err = dao.CacheFullNodeInfo(s.ctx, fullNodeInfo)
	if err != nil {
		log.Errorf("cache full node info: %v", err)
		return err
	}
	return nil
}
