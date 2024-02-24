package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/cleanup"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

var schedulerAdmin api.Scheduler

var schedulerApi api.Scheduler

//var SchedulerConfigs map[string][]*types.SchedulerCfg

var (
	DefaultAreaId            = "Asia-China-Guangdong-Shenzhen"
	SchedulerConfigKeyPrefix = "TITAN::SCHEDULERCFG"
)

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
	etcdClient      *EtcdClient
	statistic       *statistics.Statistic
	statisticCloser func()
}

func NewEtcdClient(addresses []string) (*EtcdClient, error) {
	etcd, err := etcdcli.New(addresses)
	if err != nil {
		return nil, err
	}

	etcdClient := &EtcdClient{
		cli: etcd,
		//schedulerConfigs: make(map[string][]*types.SchedulerCfg),
		configMap: make(map[string]*types.SchedulerCfg),
	}

	//if err := ec.loadSchedulerConfigs(); err != nil {
	//	return nil, err
	//}

	return etcdClient, nil
}

func (ec *EtcdClient) loadSchedulerConfigs() (map[string][]*types.SchedulerCfg, error) {
	resp, err := ec.cli.GetServers(types.NodeScheduler.String())
	if err != nil {
		return nil, err
	}

	schedulerConfigs := make(map[string][]*types.SchedulerCfg)

	for _, kv := range resp.Kvs {
		var configScheduler *types.SchedulerCfg
		err := etcdcli.SCUnmarshal(kv.Value, &configScheduler)
		if err != nil {
			return nil, err
		}
		configs, ok := schedulerConfigs[configScheduler.AreaID]
		if !ok {
			configs = make([]*types.SchedulerCfg, 0)
		}
		configs = append(configs, configScheduler)

		schedulerConfigs[configScheduler.AreaID] = configs
		ec.configMap[string(kv.Key)] = configScheduler
	}

	for areaId, cfgs := range schedulerConfigs {
		if err := SetSchedulerConfigs(context.Background(), fmt.Sprintf("%s::%s", SchedulerConfigKeyPrefix, areaId), cfgs); err != nil {
			return nil, err
		}
	}

	//ec.schedulerConfigs = schedulerConfigs
	return schedulerConfigs, nil
}

func NewServer(cfg config.Config) (*Server, error) {
	gin.SetMode(cfg.Mode)
	router := gin.Default()
	
	//router.Use(Cors())

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

func FetchSchedulersFromEtcd(etcdClient *EtcdClient) ([]*statistics.Scheduler, error) {

	schedulerConfigs, err := etcdClient.loadSchedulerConfigs()
	if err != nil {
		log.Errorf("load scheduer from etcd: %v", err)
		return nil, err
	}

	var out []*statistics.Scheduler

	for key, schedulerURLs := range schedulerConfigs {
		for _, SchedulerCfg := range schedulerURLs {
			// https protocol still in test, we use http for now.
			schedulerURL := strings.Replace(SchedulerCfg.SchedulerURL, "https", "http", 1)
			headers := http.Header{}
			headers.Add("Authorization", "Bearer "+SchedulerCfg.AccessToken)
			clientInit, closeScheduler, err := client.NewScheduler(context.Background(), schedulerURL, headers)
			if err != nil {
				log.Errorf("create scheduler rpc client: %v", err)
			}
			out = append(out, &statistics.Scheduler{
				Uuid:   schedulerURL,
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

//func (s *Server) sendEmail(sendTo string, registrations []string) error {
//	subject := "[Application]: Your Device Info"
//	contentType := "text/html"
//	content := "<h1>Your Device ID ：</h1>\n"
//	for _, registration := range registrations {
//		content += registration + "<br>"
//	}
//	port, err := strconv.ParseInt(s.cfg.Email.SMTPPort, 10, 64)
//	message := mail.NewEmailMessage(s.cfg.Email.From, subject, contentType, content, "", []string{sendTo}, nil)
//	_, err = mail.NewEmailClient(s.cfg.Email.SMTPHost, s.cfg.Email.Username, s.cfg.Email.Password, int(port), message).SendMessage()
//	if err != nil {
//		return err
//	}
//	return nil
//}

func getSchedulerClient(ctx context.Context, areaId string) (api.Scheduler, error) {
	schedulers, err := GetSchedulerConfigs(ctx, fmt.Sprintf("%s::%s", SchedulerConfigKeyPrefix, areaId))
	if err == redis.Nil && areaId != DefaultAreaId {
		return getSchedulerClient(ctx, DefaultAreaId)
	}

	if err != nil || len(schedulers) == 0 {
		log.Errorf("no scheduler found")
		return nil, errors.New("no scheduler found")
	}

	schedulerApiUrl := schedulers[0].SchedulerURL
	schedulerApiToken := schedulers[0].AccessToken
	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+schedulerApiToken)
	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
	if err != nil {
		log.Errorf("create scheduler rpc client: %v", err)
	}

	return schedulerClient, nil
}

func GetSchedulerConfigs(ctx context.Context, key string) ([]*types.SchedulerCfg, error) {
	result, err := dao.RedisCache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cfg []*types.SchedulerCfg
	err = json.Unmarshal([]byte(result), &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func SetSchedulerConfigs(ctx context.Context, key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = dao.RedisCache.Set(ctx, key, data, 0).Result()
	if err != nil {
		log.Errorf("set chain head: %v", err)
	}

	return nil
}
