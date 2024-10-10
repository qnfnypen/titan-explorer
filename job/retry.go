package job

import (
	"context"

	"github.com/hibiken/asynq"
)

type asynqHandlerhandler func(ctx context.Context, task *asynq.Task) error

// TODO retry-interval

// func retryHandlerWrapper(delayFunc asynq.RetryDelayFunc, fn asynqHandlerhandler) asynqHandlerhandler {
// 	return asynqHandlerhandler(func(ctx context.Context, task *asynq.Task) error {
// 		err := fn(ctx, task)
// 		if err != nil {
// 			task.Payload()
// 			delayFunc()
// 		}
// 		return nil
// 	})
// }
