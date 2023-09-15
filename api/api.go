package api

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/backup"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/utils"
)

var schedulerAdmin api.Scheduler

var schedulerApi api.Scheduler

var ApplicationC chan bool

var SchedulerConfigs map[string][]*types.SchedulerCfg

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
	schedulerClient api.Scheduler
	etcdClient      *EtcdClient
	statistic       *statistics.Statistic
	statisticCloser func()
	storageBackup   *backup.StorageBackup
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

	SchedulerConfigs = make(map[string][]*types.SchedulerCfg)

	for _, kv := range resp.Kvs {
		var configScheduler *types.SchedulerCfg
		err := etcdcli.SCUnmarshal(kv.Value, &configScheduler)
		if err != nil {
			return err
		}
		configs, ok := SchedulerConfigs[configScheduler.AreaID]
		if !ok {
			configs = make([]*types.SchedulerCfg, 0)
		}
		configs = append(configs, configScheduler)

		SchedulerConfigs[configScheduler.AreaID] = configs
		ec.configMap[string(kv.Key)] = configScheduler
	}

	ec.schedulerConfigs = SchedulerConfigs
	return nil
}

func NewServer(cfg config.Config) (*Server, error) {
	gin.SetMode(cfg.Mode)
	router := gin.Default()
	//router.Use(cors.New(cors.Config{
	//	//AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	//	//AllowHeaders:     []string{"Authorization", "Origin", "Content-Length", "Content-Type"},
	//	//AllowCredentials: true,
	//	MaxAge: 12 * time.Hour,
	//	//AllowAllOrigins:  true,
	//}))
	ConfigRouter(router, cfg)
	var address []string
	address = append(address, cfg.EtcdAddress)
	eClient, err := NewEtcdClient(address)
	if err != nil {
		log.Errorf("New etcdClient Failed: %v", err)
		return nil, err
	}

	schedulers, err := fetchSchedulersFromEtcd(eClient)
	if err != nil {
		log.Errorf("fetch scheduler from etcd Failed: %v", err)
		return nil, err
	}

	if cfg.AdminScheduler.Enable {
		applyAdminScheduler(cfg.AdminScheduler.Address, cfg.AdminScheduler.Token)
	}
	s := &Server{
		cfg:           cfg,
		router:        router,
		statistic:     statistics.New(cfg.Statistic, schedulers),
		etcdClient:    eClient,
		storageBackup: backup.NewStorageBackup(cfg.StorageBackup, schedulers),
	}

	return s, nil
}

func (s *Server) Run() {
	s.statistic.Run()
	s.asyncHandleApplication()
	s.storageBackup.Run()
	err := s.router.Run(s.cfg.ApiListen)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Close() {
	s.statistic.Stop()
}

func fetchSchedulersFromEtcd(locatorApi *EtcdClient) ([]*statistics.Scheduler, error) {
	var out []*statistics.Scheduler
	for key, SchedulerURLs := range locatorApi.schedulerConfigs {
		for _, SchedulerCfg := range SchedulerURLs {
			// https protocol still in test, we use http for now.
			SchedulerURL := strings.Replace(SchedulerCfg.SchedulerURL, "https", "http", 1)
			headers := http.Header{}
			headers.Add("Authorization", "Bearer "+SchedulerCfg.AccessToken)
			clientInit, closeScheduler, err := client.NewScheduler(context.Background(), SchedulerURL, headers)
			if err != nil {
				log.Errorf("create scheduler rpc client: %v", err)
			}
			out = append(out, &statistics.Scheduler{
				Uuid:   SchedulerURL,
				Api:    clientInit,
				AreaId: key,
				Closer: closeScheduler,
			})
			schedulerApi = clientInit

		}

	}

	log.Infof("fetch %d schedulers from Etcd", len(out))

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

func (s *Server) asyncHandleApplication() {
	ctx := context.Background()
	handleApplicationInterval := 5 * time.Minute
	ticker := time.NewTicker(handleApplicationInterval)
	ApplicationC = make(chan bool, 1)
	reSendEmailInterval := time.Minute
	reSendEmailTicker := time.NewTicker(reSendEmailInterval)

	i := rand.Int()
	go func() {
		for {
			select {
			case <-ticker.C:
				ApplicationC <- true
			case <-ApplicationC:
				applications, err := dao.GetApplicationList(ctx, []int{
					dao.ApplicationStatusCreated,
					dao.ApplicationStatusFailed,
				})
				if err != nil {
					log.Errorf("get applications: %v", err)
					continue
				}

				for _, application := range applications {
					selectedScheduler := s.etcdClient.schedulerConfigs[application.AreaID]
					if len(selectedScheduler) < 1 {
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
					err = s.handleApplication(ctx, application.PublicKey, application, int(application.Amount))
					if err != nil {
						log.Errorf("handleApplication %v", err)
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
					//err = s.sendEmail(application.Email, infos)
					//if err != nil {
					//	status = dao.ApplicationStatusSendEmailFailed
					//	log.Errorf("send email appicationID:%d, %v", application.ID, err)
					//}

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

func (s *Server) handleApplication(ctx context.Context, publicKey string, application *model.Application, amount int) error {
	registration, err := s.schedulerClient.RequestActivationCodes(ctx, types.NodeType(1), amount)
	if err != nil {
		log.Errorf("register node: %v", err)
		return err
	}
	var results []*model.ApplicationResult
	var deviceInfos []*model.DeviceInfo
	for _, deviceInfo := range registration {
		results = append(results, &model.ApplicationResult{
			UserID:        application.UserID,
			DeviceID:      deviceInfo.NodeID,
			NodeType:      1,
			ApplicationID: application.ID,
			Secret:        deviceInfo.ActivationCode,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		deviceInfos = append(deviceInfos, &model.DeviceInfo{
			UserID:       application.UserID,
			IpLocation:   application.AreaID,
			DeviceID:     deviceInfo.NodeID,
			NodeType:     1,
			DeviceName:   "",
			BindStatus:   "binding",
			ActiveStatus: 0,
		})
	}

	if len(results) == 0 {
		log.Infof("update application status: %v", registration)
		return nil
	}

	status := dao.ApplicationStatusFinished
	defer func() {
		er := dao.UpdateApplicationStatus(ctx, application.ID, status)
		if er != nil {
			log.Errorf("update application status: %v", err)
		}
	}()

	e := dao.AddApplicationResult(ctx, results)
	if e != nil {
		status = dao.ApplicationStatusFailed
		log.Errorf("create application result: %v", err)
		return e
	}

	e = dao.BulkUpsertDeviceInfo(ctx, deviceInfos)
	if e != nil {
		log.Errorf("add device info: %v", err)
	}
	//var registrations []string
	//err = s.sendEmail(application.Email, append(registrations, registration))
	//if err != nil {
	//	status = dao.ApplicationStatusSendEmailFailed
	//	log.Errorf("send email appicationID:%d, %v", application.ID, err)
	//	return err
	//}

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

func GetNewScheduler(ctx context.Context, areaId string) api.Scheduler {
	scheduler, _ := SchedulerConfigs[areaId]
	if len(scheduler) < 1 {
		scheduler = SchedulerConfigs["Asia-China-Guangdong-Shenzhen"]
	}
	schedulerApiUrl := scheduler[0].SchedulerURL
	schedulerApiToken := scheduler[0].AccessToken
	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+schedulerApiToken)
	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}
	return schedulerClient
}
