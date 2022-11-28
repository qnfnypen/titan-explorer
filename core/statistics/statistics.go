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
	s.cron.AddFunc("0 * * * * *", s.UpdateDeviceInfo)
	s.cron.AddFunc("0 */1 * * * *", s.StatFullNodeInfoByMinutes)
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

const LockerTTL = 30 * time.Second

func (s *Statistic) Once(ctx context.Context, key string, fn func() error) error {
	lock, err := s.locker.Obtain(ctx, key, LockerTTL, nil)
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
