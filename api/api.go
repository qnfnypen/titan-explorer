package api

import (
	"context"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/cleanup"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/pkg/mail"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

var schedulerAdmin api.Scheduler

var schedulerApi api.Scheduler

//var ApplicationC chan bool

var SchedulerConfigs map[string][]*types.SchedulerCfg

type EtcdClient struct {
	cli *etcdcli.Client
	// key is areaID, value is array of types.SchedulerCfg pointer
	schedulerConfigs map[string][]*types.SchedulerCfg
	// key is etcd key, value is types.SchedulerCfg pointer
	configMap map[string]*types.SchedulerCfg
}

type Server struct {
	cfg    config.Config
	router *gin.Engine
	//schedulerClient api.Scheduler
	etcdClient      *EtcdClient
	statistic       *statistics.Statistic
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

	router.Use(Cors())

	RegisterRouters(router, cfg)

	etcdClient, err := NewEtcdClient([]string{cfg.EtcdAddress})
	if err != nil {
		log.Errorf("New etcdClient Failed: %v", err)
		return nil, err
	}

	schedulers, err := FetchSchedulersFromEtcd(etcdClient)
	if err != nil {
		log.Errorf("fetch scheduler from etcd Failed: %v", err)
		return nil, err
	}

	if cfg.AdminScheduler.Enable {
		applyAdminScheduler(cfg.AdminScheduler.Address, cfg.AdminScheduler.Token)
	}
	s := &Server{
		cfg:        cfg,
		router:     router,
		statistic:  statistics.New(cfg.Statistic, schedulers),
		etcdClient: etcdClient,
	}

	go cleanup.Run(context.Background())

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
	s.statistic.Stop()
}

func FetchSchedulersFromEtcd(locatorApi *EtcdClient) ([]*statistics.Scheduler, error) {
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
	schedClient, _, err := client.NewScheduler(context.Background(), url, headers)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}
	schedulerAdmin = schedClient
}

func (s *Server) sendEmail(sendTo string, registrations []string) error {
	subject := "[Application]: Your Device Info"
	contentType := "text/html"
	content := "<h1>Your Device ID ：</h1>\n"
	for _, registration := range registrations {
		content += registration + "<br>"
	}
	port, err := strconv.ParseInt(s.cfg.Email.SMTPPort, 10, 64)
	message := mail.NewEmailMessage(s.cfg.Email.From, subject, contentType, content, "", []string{sendTo}, nil)
	_, err = mail.NewEmailClient(s.cfg.Email.SMTPHost, s.cfg.Email.Username, s.cfg.Email.Password, int(port), message).SendMessage()
	if err != nil {
		return err
	}
	return nil
}

func getSchedulerClient(ctx context.Context, areaId string) (api.Scheduler, error) {
	scheduler, _ := SchedulerConfigs[areaId]
	if len(scheduler) < 1 {
		scheduler = SchedulerConfigs["Asia-China-Guangdong-Shenzhen"]
	}

	if len(scheduler) == 0 {
		log.Errorf("no scheduler found")
		return nil, errors.New("no scheduler found")
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

	return schedulerClient, nil
}
