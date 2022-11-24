package api

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/linguohua/titan/api"
	"github.com/linguohua/titan/api/client"
)

var schedulerClient api.Scheduler

type Server struct {
	cfg             config.Config
	router          *gin.Engine
	schedulerClient api.Scheduler
	statistic       *statistics.Statistic
	closer          func()
}

func NewServer(cfg config.Config) (*Server, error) {
	gin.SetMode(cfg.Mode)
	router := gin.Default()
	router.Use(cors.Default())
	ConfigRouter(router, cfg)

	client, closer, err := getSchedulerClient()
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:             cfg,
		router:          router,
		statistic:       statistics.New(client),
		schedulerClient: client,
		closer:          func() { closer() },
	}

	go RunTask()

	return s, nil
}

func (s *Server) Run() {
	s.statistic.Run()

	err := s.router.Run(s.cfg.ApiListen)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Close() {
	select {
	case <-s.statistic.Stop().Done():
	}
	s.closer()
}

func getSchedulerClient() (api.Scheduler, func(), error) {
	schedulers, err := dao.GetSchedulers(context.Background())
	if err != nil {
		return nil, nil, err
	}

	if len(schedulers) == 0 {
		log.Fatalf("scheulers not found")
	}

	addr := schedulers[0].Address
	client, closeScheduler, err := client.NewScheduler(context.Background(), addr, nil)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}

	schedulerClient = client
	return client, closeScheduler, nil
}
