package statistics

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/gnasnik/titan-explorer/core/dao"
	logging "github.com/ipfs/go-log/v2"
	"github.com/linguohua/titan/api"
	"github.com/robfig/cron/v3"
	"time"
)

var log = logging.Logger("statistics")

const LockerTTL = 30 * time.Second

const DKeyFetchAllNodes = "titan::dk_fetch_all_nodes"

func (s *Statistic) initContabs() {
	s.cron.AddFunc("@every 1m", s.Once(DKeyFetchAllNodes, s.FetchAllNodes))
}

type Statistic struct {
	cron   *cron.Cron
	api    api.Scheduler
	locker *redislock.Client
}

func New(api api.Scheduler) *Statistic {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(time.Local),
	)

	s := &Statistic{
		api:    api,
		cron:   c,
		locker: redislock.New(dao.Cache),
	}

	s.initContabs()

	return s
}

func (s *Statistic) Run() {
	s.cron.Start()
}

func (s *Statistic) Stop() context.Context {
	return s.cron.Stop()
}

func (s *Statistic) Once(key string, fn func() error) func() {
	return func() {
		ctx := context.Background()
		lock, err := s.locker.Obtain(ctx, key, LockerTTL, nil)
		if err == redislock.ErrNotObtained {
			log.Debug(redislock.ErrNotObtained)
			return
		}

		if err != nil {
			log.Fatalf("obtain redis lock: %v", err)
			return
		}

		defer lock.Release(ctx)

		if err = fn(); err != nil {
			log.Errorf("execute cron job: %v", err)
		}
	}
}
