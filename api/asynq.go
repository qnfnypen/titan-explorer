package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

func newMonitorHandlerByRedisCfg(addr, pass string) gin.HandlerFunc {
	return gin.WrapH(asynqmon.New(asynqmon.Options{
		RootPath:     "/api/v1/monitor",
		RedisConnOpt: asynq.RedisClientOpt{Addr: addr, Password: pass},
	}))
}
