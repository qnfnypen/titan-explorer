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

type Statistic struct {
	cron   *cron.Cron
	api    api.Scheduler
	locker *redislock.Client
}

func (s *Statistic) initContabs() {
	s.cron.AddFunc("1 * * * * *", s.UpdateDeviceInfo)
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

func (s *Statistic) Once(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
	lock, err := s.locker.Obtain(ctx, key, expiration, nil)
	if err == redislock.ErrNotObtained {
		log.Debug(redislock.ErrNotObtained)
		return nil
	}

	if err != nil {
		log.Fatalf("obtain redis lock: %v", err)
		return err
	}

	defer lock.Release(ctx)

	return fn()
}
