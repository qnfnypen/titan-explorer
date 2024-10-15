package api

import (
	"github.com/gnasnik/titan-explorer/config"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

var MonitorHandler = asynqmon.New(asynqmon.Options{
	RootPath:     "/api/v1/monitor",
	RedisConnOpt: asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword},
})
