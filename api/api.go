package api

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/utils"
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
	ctx := context.Background()
	handleApplicationInterval := time.Minute
	ticker := time.NewTicker(handleApplicationInterval)

	reSendEmailInterval := time.Minute
	reSendEmailTicker := time.NewTicker(reSendEmailInterval)

	for {
		select {
		case <-ticker.C:
			applications, err := dao.GetApplicationList(ctx, []int{
				dao.ApplicationStatusCreated,
				dao.ApplicationStatusFailed,
			})
			if err != nil {
				log.Errorf("get applications: %v", err)
				continue
			}

			for _, application := range applications {
				s.handleApplication(ctx, application)
			}

			ticker.Reset(handleApplicationInterval)
		case <-reSendEmailTicker.C:
			applications, err := dao.GetApplicationList(ctx, []int{
				dao.ApplicationStatusSendEmailFailed,
			})
			if err != nil {
				log.Errorf("get applications: %v", err)
				continue
			}

			for _, application := range applications {
				results, err := dao.GetApplicationResults(ctx, application.ID)
				if err != nil {
					log.Errorf("get application results: %v", err)
					continue
				}

				var infos []api.NodeRegisterInfo
				for _, res := range results {
					infos = append(infos, api.NodeRegisterInfo{
						DeviceID: res.DeviceID,
						Secret:   res.Secret,
					})
				}

				status := dao.ApplicationStatusFinished
				err = s.sendEmail(application.Email, infos)
				if err != nil {
					status = dao.ApplicationStatusSendEmailFailed
					log.Errorf("send email appicationID:%d, %v", application.ID, err)
				}

				err = dao.UpdateApplicationStatus(ctx, application.ID, status)
				if err != nil {
					log.Errorf("update application status: %v", err)
				}
			}
			reSendEmailTicker.Reset(reSendEmailInterval)
		case <-ctx.Done():
			return
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
	var deviceInfos []*model.DeviceInfo
	for _, registration := range registrations {
		results = append(results, &model.ApplicationResult{
			UserID:        application.UserID,
			DeviceID:      registration.DeviceID,
			NodeType:      application.NodeType,
			Secret:        registration.Secret,
			ApplicationID: application.ID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		deviceInfos = append(deviceInfos, &model.DeviceInfo{
			UserID:   application.UserID,
			DeviceID: registration.DeviceID,
		})
	}

	if len(results) == 0 {
		return nil
	}

	status := dao.ApplicationStatusFinished
	defer func() {
		err = dao.UpdateApplicationStatus(ctx, application.ID, status)
		if err != nil {
			log.Errorf("update application status: %v", err)
		}
	}()

	err = dao.AddApplicationResult(ctx, results)
	if err != nil {
		status = dao.ApplicationStatusFailed
		log.Errorf("create application result: %v", err)
		return err
	}

	err = dao.BulkUpsertDeviceInfo(ctx, deviceInfos)
	if err != nil {
		log.Errorf("add device info: %v", err)
	}

	err = s.sendEmail(application.Email, registrations)
	if err != nil {
		status = dao.ApplicationStatusSendEmailFailed
		log.Errorf("send email appicationID:%d, %v", application.ID, err)
		return err
	}

	return nil
}

func (s *Server) sendEmail(sendTo string, results []api.NodeRegisterInfo) error {
	var EData utils.EmailData
	EData.Subject = "[Application]: Your Device Info"
	EData.Tittle = "please check your device id and secret"
	EData.SendTo = sendTo
	EData.Content = "<h1>Your Device ID and Secret：</h1>\n"
	for _, registration := range results {
		EData.Content += registration.DeviceID + ":" + registration.Secret + "<br>"
	}
	err := utils.SendEmail(s.cfg.Email, EData)
	if err != nil {
		return err
	}
	return nil
}
