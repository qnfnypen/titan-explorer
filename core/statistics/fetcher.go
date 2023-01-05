package statistics

import (
	"context"
	"github.com/linguohua/titan/api"
)

type Fetcher interface {
	Fetch(ctx context.Context, scheduler *Scheduler) error
	Push(ctx context.Context, job Job)
	GetJobQueue() chan Job
}

type Scheduler struct {
	Uuid   string
	Api    api.Scheduler
	Closer func()
}

type Job func() error
