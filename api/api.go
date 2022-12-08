package api

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/linguohua/titan/api"
	"github.com/linguohua/titan/api/client"
	"time"
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

	return s, nil
}

func (s *Server) Run() {
	s.statistic.Run()

	go s.asyncHandleApplication()

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

func (s *Server) asyncHandleApplication() {
	handleApplicationInterval := time.Minute
	ticker := time.NewTicker(handleApplicationInterval)
	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			applications, err := dao.GetApplicationList(ctx)
			if err != nil {
				log.Errorf("get applications: %v", err)
				continue
			}

			for _, application := range applications {
				s.handleApplication(ctx, application)
			}

			ticker.Reset(handleApplicationInterval)
		}
	}
}

func (s *Server) handleApplication(ctx context.Context, application *model.Application) error {
	registrations, err := s.schedulerClient.RegisterNode(ctx, api.NodeType(application.NodeType), int(application.Amount))
	if err != nil {
		log.Errorf("register node: %v", err)
		return err
	}

	var results []*model.ApplicationResult
	for _, registration := range registrations {
		results = append(results, &model.ApplicationResult{
			UserID:    application.UserID,
			DeviceID:  registration.DeviceID,
			NodeType:  application.NodeType,
			Secret:    registration.Secret,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	if len(results) == 0 {
		return nil
	}

	err = dao.AddApplicationResult(ctx, results)
	if err != nil {
		log.Errorf("create application result: %v", err)

		err = dao.UpdateApplicationStatus(ctx, application.ID, dao.ApplicationStatusFailed)
		if err != nil {
			log.Errorf("update application status: %v", err)
			return err
		}

		return err
	}

	err = dao.UpdateApplicationStatus(ctx, application.ID, dao.ApplicationStatusSuccess)
	if err != nil {
		log.Errorf("update application status: %v", err)
		return err
	}

	return nil
}
