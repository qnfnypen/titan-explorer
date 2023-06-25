package statistics

import (
	"fmt"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/bsm/redislock"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/context"
	"net/http"
	"strings"
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
const statisticLockerKeyPrefix = "TITAN::STATISTIC"

var SchedulerConfigs map[string][]*types.SchedulerCfg

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
			//newCacheFetcher(),
			//newRetrievalFetcher(),
			//newValidationFetcher(),
			newSystemInfoFetcher(),
			newStorageFetcher(),
		},
	}

	return s
}

func (s *Statistic) Run() {
	if s.cfg.Disable {
		return
	}
	s.cron.AddFunc(s.cfg.Crontab, s.Once("FETCHER", s.runFetchers))
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

func (s *Statistic) runFetchers() error {
	var wg sync.WaitGroup
	wg.Add(len(s.schedulers))
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
	wg.Wait()

	s.asyncExecute([]func() error{
		s.SumDeviceInfoProfit,
		//s.CountRetrievals,
		s.SumAllNodes,
		s.UpdateDeviceRank,
	})

	return nil
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

func GetNewScheduler(ctx context.Context, areaId string) api.Scheduler {
	scheduler, _ := SchedulerConfigs[areaId]
	if len(scheduler) < 1 {
		scheduler = SchedulerConfigs["Asia-China-Guangdong-Shenzhen"]
	}
	schedulerApiUrl := scheduler[0].SchedulerURL
	schedulerApiToken := scheduler[0].AccessToken
	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+schedulerApiToken)
	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}
	return schedulerClient
}
