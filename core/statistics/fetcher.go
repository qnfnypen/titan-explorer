package statistics

import (
	"context"
	"github.com/Filecoin-Titan/titan/api"
)

type Fetcher interface {
	Fetch(ctx context.Context, scheduler *Scheduler) error
	Push(ctx context.Context, job Job)
	GetJobQueue() chan Job
}

type Scheduler struct {
	Uuid   string
	AreaId string
	Api    api.Scheduler
	Closer func()
}

type Job func() error

type BaseFetcher struct {
	jobQueue chan Job
}

func newBaseFetcher() BaseFetcher {
	return BaseFetcher{jobQueue: make(chan Job, 1)}
}

func (b BaseFetcher) Push(ctx context.Context, job Job) {
	select {
	case b.jobQueue <- job:
	case <-ctx.Done():
	}
}

func (b BaseFetcher) GetJobQueue() chan Job {
	return b.jobQueue
}
