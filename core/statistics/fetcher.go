package statistics

import (
	"context"
	"github.com/Filecoin-Titan/titan/api"
)

// Fetcher is an interface for fetching and processing data.
type Fetcher interface {
	Fetch(ctx context.Context, scheduler *Scheduler) error
	Push(ctx context.Context, job Job)
	GetJobQueue() chan Job
	Finalize() error
}

type Scheduler struct {
	Uuid   string
	AreaId string
	Api    api.Scheduler
	Closer func()
}

// Job is a function that can be executed as a job.
type Job func() error

// BaseFetcher is a basic implementation of the Fetcher interface.
type BaseFetcher struct {
	jobQueue chan Job
}

// newBaseFetcher creates a new BaseFetcher instance.
func newBaseFetcher() BaseFetcher {
	return BaseFetcher{jobQueue: make(chan Job, 100)}
}

// Push adds a job to the job queue, respecting the context.
func (b BaseFetcher) Push(ctx context.Context, job Job) {
	select {
	case b.jobQueue <- job:
	case <-ctx.Done():
	}
}

// GetJobQueue returns the job queue channel.
func (b BaseFetcher) GetJobQueue() chan Job {
	return b.jobQueue
}
