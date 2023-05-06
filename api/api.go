package api

import (
	"context"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/utils"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var schedulerAdmin api.Scheduler

type EtcdClient struct {
	cli *etcdcli.Client
	// key is areaID, value is array of types.SchedulerCfg pointer
	schedulerConfigs map[string][]*types.SchedulerCfg
	// key is etcd key, value is types.SchedulerCfg pointer
	configMap map[string]*types.SchedulerCfg
}

type Server struct {
	cfg             config.Config
	router          *gin.Engine
	locatorClient   api.Locator
	schedulerClient api.Scheduler
	etcdClient      *EtcdClient
	statistic       *statistics.Statistic
	locatorCloser   func()
	statisticCloser func()
}

func NewEtcdClient(addresses []string) (*EtcdClient, error) {
	etcd, err := etcdcli.New(addresses)
	if err != nil {
		return nil, err
	}

	ec := &EtcdClient{
		cli:              etcd,
		schedulerConfigs: make(map[string][]*types.SchedulerCfg),
		configMap:        make(map[string]*types.SchedulerCfg),
	}

	if err := ec.loadSchedulerConfigs(); err != nil {
		return nil, err
	}

	return ec, nil
}

func (ec *EtcdClient) loadSchedulerConfigs() error {
	resp, err := ec.cli.GetServers(types.NodeScheduler.String())
	if err != nil {
		return err
	}

	schedulerConfigs := make(map[string][]*types.SchedulerCfg)

	for _, kv := range resp.Kvs {
		var configScheduler *types.SchedulerCfg
		err := etcdcli.SCUnmarshal(kv.Value, &configScheduler)
		if err != nil {
			return err
		}

		configs, ok := schedulerConfigs[configScheduler.AreaID]
		if !ok {
			configs = make([]*types.SchedulerCfg, 0)
		}
		configs = append(configs, configScheduler)

		schedulerConfigs[configScheduler.AreaID] = configs
		ec.configMap[string(kv.Key)] = configScheduler
	}

	ec.schedulerConfigs = schedulerConfigs
	return nil
}

func NewServer(cfg config.Config) (*Server, error) {
	gin.SetMode(cfg.Mode)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowAllOrigins:  true,
	}))
	ConfigRouter(router, cfg)

	client, closer, err := getLocatorClient(cfg.Locator.Address, cfg.Locator.Token)
	if err != nil {
		return nil, err
	}
	version, err := client.Version(context.Background())
	if err != nil {
		log.Errorf("get version from locator: %v", err)
		return nil, err
	}

	log.Infof("Locator connected, url: %s, version: %s", cfg.Locator.Address, version)

	var schedulers []*statistics.Scheduler
	if cfg.SchedulerFromDB {
		schedulers, err = fetchSchedulersFromDatabase()
		if err != nil {
			return nil, err
		}
	} else {
		schedulers, err = fetchSchedulersFromLocator(client)
		if err != nil {
			return nil, err
		}
	}

	var address []string
	// todo
	address = append(address, "39.108.143.56:2379")
	eClient, err := NewEtcdClient(address)
	if err != nil {
		log.Errorf("New etcdClient Failed: %v", err)
		return nil, err
	}
	schedulers, err = fetchSchedulersFromEtcd(eClient)
	if cfg.AdminScheduler.Enable {
		applyAdminScheduler(cfg.AdminScheduler.Address, cfg.AdminScheduler.Token)
	}
	s := &Server{
		cfg:           cfg,
		router:        router,
		statistic:     statistics.New(cfg.Statistic, schedulers),
		locatorClient: client,
		etcdClient:    eClient,
		locatorCloser: closer,
	}

	return s, nil
}

func (s *Server) Run() {
	s.statistic.Run()
	s.asyncHandleApplication()
	err := s.router.Run(s.cfg.ApiListen)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Close() {
	s.locatorCloser()
	s.statistic.Stop()
}

func fetchSchedulersFromDatabase() ([]*statistics.Scheduler, error) {
	schedulers, err := dao.GetSchedulers(context.Background())
	if err != nil {
		return nil, err
	}

	if len(schedulers) == 0 {
		log.Fatalf("scheulers not found")
	}

	var out []*statistics.Scheduler
	for _, item := range schedulers {
		// read permission only
		client, closeScheduler, err := client.NewScheduler(context.Background(), item.Address, nil)
		if err != nil {
			log.Errorf("create scheduler rpc client: %v", err)
		}
		out = append(out, &statistics.Scheduler{
			Uuid:   item.Uuid,
			Api:    client,
			Closer: closeScheduler,
		})
	}

	log.Infof("fetch %d schedulers from database", len(out))

	return out, nil
}

func fetchSchedulersFromLocator(locatorApi api.Locator) ([]*statistics.Scheduler, error) {
	accessPoints, err := locatorApi.GetUserAccessPoint(context.Background(), "")
	if err != nil {
		log.Errorf("api GetAccessPoints: %v", err)
		return nil, err
	}

	var out []*statistics.Scheduler
	// todo etcd
	for _, SchedulerURL := range accessPoints.SchedulerURLs {
		// https protocol still in test, we use http for now.
		SchedulerURL = strings.Replace(SchedulerURL, "https", "http", 1)
		client, closeScheduler, err := client.NewScheduler(context.Background(), SchedulerURL, nil)
		if err != nil {
			log.Errorf("create scheduler rpc client: %v", err)
		}
		out = append(out, &statistics.Scheduler{
			Uuid:   SchedulerURL,
			Api:    client,
			Closer: closeScheduler,
		})

	}

	log.Infof("fetch %d schedulers from locator", len(out))

	return out, nil
}

func fetchSchedulersFromEtcd(locatorApi *EtcdClient) ([]*statistics.Scheduler, error) {
	var out []*statistics.Scheduler
	for key, SchedulerURLs := range locatorApi.schedulerConfigs {
		for _, SchedulerCfg := range SchedulerURLs {
			// https protocol still in test, we use http for now.
			SchedulerURL := strings.Replace(SchedulerCfg.SchedulerURL, "https", "http", 1)
			headers := http.Header{}
			headers.Add("Authorization", "Bearer "+SchedulerCfg.AccessToken)
			client, closeScheduler, err := client.NewScheduler(context.Background(), SchedulerURL, headers)
			if err != nil {
				log.Errorf("create scheduler rpc client: %v", err)
			}
			out = append(out, &statistics.Scheduler{
				Uuid:   SchedulerURL,
				Api:    client,
				AreaId: key,
				Closer: closeScheduler,
			})

		}

	}

	log.Infof("fetch %d schedulers from locator", len(out))

	return out, nil
}

func applyAdminScheduler(url string, token string) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	client, _, err := client.NewScheduler(context.Background(), url, headers)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}
	schedulerAdmin = client
}

func getLocatorClient(address, token string) (api.Locator, func(), error) {
	//headers := http.Header{}
	//headers.Add("Authorization", "Bearer "+string(token))
	client, closer, err := client.NewLocator(context.Background(), address, nil)
	if err != nil {
		log.Errorf("create locator rpc client: %v", err)
		return nil, nil, err
	}

	return client, closer, nil
}

func (s *Server) asyncHandleApplication() {
	ctx := context.Background()
	handleApplicationInterval := time.Minute
	ticker := time.NewTicker(handleApplicationInterval)

	reSendEmailInterval := time.Minute
	reSendEmailTicker := time.NewTicker(reSendEmailInterval)

	i := rand.Int()

	go func() {
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
					//accessPoints, err := s.locatorClient.GetUserAccessPoint(ctx, application.Ip)
					//if err != nil {
					//	log.Errorf("get access points: %v", err)
					//	continue
					//}
					//
					//if len(accessPoints.SchedulerURLs) == 0 {
					//	log.Error("no accessPoints schedulerInfos return")
					//	continue
					//}

					selectedScheduler := s.etcdClient.schedulerConfigs[application.AreaID]
					if len(selectedScheduler) < 1 {
						log.Errorf("get no scheduler by this area id: %v", application.AreaID)
						continue
					}
					i++
					SchedulerURL := strings.Replace(selectedScheduler[i%len(selectedScheduler)].SchedulerURL, "https", "http", 1)
					headers := http.Header{}
					headers.Add("Authorization", "Bearer "+selectedScheduler[i%len(selectedScheduler)].AccessToken)
					schedulerClient, _, err := client.NewScheduler(context.Background(), SchedulerURL, headers)
					if err != nil {
						log.Errorf("create scheduler rpc client: %v", err)
						continue
					}
					s.schedulerClient = schedulerClient
					err = s.handleApplication(ctx, "", application)
					if err != nil {
						continue
					}
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

					var infos []string
					for _, res := range results {
						infos = append(infos, res.DeviceID)
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
	}()
}

func (s *Server) handleApplication(ctx context.Context, publicKey string, application *model.Application) error {
	registration, err := s.schedulerClient.RegisterNode(ctx, publicKey, types.NodeType(application.NodeType))
	if err != nil {
		log.Errorf("register node: %v", err)
		return err
	}
	var results []*model.ApplicationResult
	var deviceInfos []*model.DeviceInfo
	results = append(results, &model.ApplicationResult{
		UserID:        application.UserID,
		DeviceID:      registration,
		NodeType:      application.NodeType,
		ApplicationID: application.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	deviceInfos = append(deviceInfos, &model.DeviceInfo{
		UserID:     application.UserID,
		IpLocation: application.AreaID,
		DeviceID:   registration,
	})

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
	var registrations []string
	err = s.sendEmail(application.Email, append(registrations, registration))
	if err != nil {
		status = dao.ApplicationStatusSendEmailFailed
		log.Errorf("send email appicationID:%d, %v", application.ID, err)
		return err
	}

	return nil
}

func (s *Server) sendEmail(sendTo string, registrations []string) error {
	var EData utils.EmailData
	EData.Subject = "[Application]: Your Device Info"
	EData.Tittle = "please check your device id "
	EData.SendTo = sendTo
	EData.Content = "<h1>Your Device ID ：</h1>\n"
	for _, registration := range registrations {
		EData.Content += registration + "<br>"
	}

	err := utils.SendEmail(s.cfg.Email, EData)
	if err != nil {
		return err
	}
	return nil
}
