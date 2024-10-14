package job

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

type asynqHandlerhandler func(ctx context.Context, task *asynq.Task) error

// TODO retry-interval

//	func retryHandlerWrapper(delayFunc asynq.RetryDelayFunc, fn asynqHandlerhandler) asynqHandlerhandler {
//		return asynqHandlerhandler(func(ctx context.Context, task *asynq.Task) error {
//			err := fn(ctx, task)
//			if err != nil {
//				task.Payload()
//				delayFunc()
//			}
//			return nil
//		})
//	}
var tenantCallbackInterval = []time.Duration{
	10 * time.Second, 30 * time.Second, 1 * time.Minute, 5 * time.Minute, 20 * time.Minute, 1 * time.Hour, 24 * time.Hour,
}

func retry(n int, e error, t *asynq.Task) time.Duration {
	if len(tenantCallbackInterval) > n {
		return tenantCallbackInterval[n]
	}
	return 24 * time.Hour
}
