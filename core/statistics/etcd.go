package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gnasnik/titan-explorer/core/dao"
)

var (
	DefaultAreaId            = "Asia-China-Guangdong-Shenzhen"
	SchedulerConfigKeyPrefix = "TITAN::SCHEDULERCFG"
)

type EtcdClient struct {
	cli *etcdcli.Client
	// key is etcd key, value is types.SchedulerCfg pointer
	configMap map[string]*types.SchedulerCfg
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

// func LoadSchedulerConfigs() (map[string][]*types.SchedulerCfg, error) {
// 	cli, err := NewEtcdClient([]string{"47.236.228.34:2379", "8.211.44.79:2379", "8.209.205.85:2379"})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return cli.loadSchedulerConfigs()
// }

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

func FetchSchedulersFromEtcd(etcdClient *EtcdClient) ([]*Scheduler, error) {

	schedulerConfigs, err := etcdClient.loadSchedulerConfigs()
	if err != nil {
		log.Errorf("load scheduer from etcd: %v", err)
		return nil, err
	}

	var out []*Scheduler

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
			out = append(out, &Scheduler{
				Uuid:   schedulerURL,
				Api:    clientInit,
				AreaId: key,
				Closer: closeScheduler,
			})
			//schedulerApi = clientInit
		}
	}

	log.Infof("fetch %d schedulers from Etcd", len(out))

	return out, nil
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
