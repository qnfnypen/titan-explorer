package statistics

import (
	"context"
	logging "github.com/ipfs/go-log/v2"
	"github.com/linguohua/titan/api"
	"github.com/robfig/cron/v3"
	"time"
)

var log = logging.Logger("statistics")

type Statistic struct {
	cron *cron.Cron
	api  api.Scheduler
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
		api:  api,
		cron: c,
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

func (s *Statistic) once(ctx context.Context, key string, expiration time.Duration, callback func() error) error {
	return callback()
}
