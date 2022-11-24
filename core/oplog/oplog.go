package oplog

import (
	"context"
	"github.com/filecoin-project/pubsub"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("oplog")

const (
	loggerLoginTopic     string = "login"
	loggerOperationTopic        = "operation"
)

var o *oplog

func init() {
	o = &oplog{
		logger: pubsub.New(50),
	}
	o.sub(context.Background())
}

type oplog struct {
	logger *pubsub.PubSub
}

func (l *oplog) pub(topic string, v interface{}) {
	l.logger.Pub(v, topic)
}

func (l *oplog) sub(ctx context.Context) {
	login := l.logger.Sub(loggerLoginTopic)
	operator := l.logger.Sub(loggerOperationTopic)
	go func() {
		defer l.logger.Unsub(login)
		defer l.logger.Unsub(operator)

		for {
			select {
			case msg := <-login:
				err := dao.AddLoginLog(ctx, msg.(*model.LoginLog))
				if err != nil {
					log.Errorf("add login log: %v", err)
				}
			case msg := <-operator:
				err := dao.AddOperationLog(ctx, msg.(*model.OperationLog))
				if err != nil {
					log.Errorf("add operation log: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func AddLoginLog(v interface{}) {
	o.pub(loggerLoginTopic, v)
}

func AddOperationLog(v interface{}) {
	o.pub(loggerOperationTopic, v)
}

func Subscribe(ctx context.Context) {
	o.sub(ctx)
}
