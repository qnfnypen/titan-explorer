package oprds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/go-redis/redis/v9"
)

const (
	key = "sync_scheduler_list"
)

var cli *Client

// Client 客户端
type Client struct {
	rds *redis.Client
}

// Payload 载体
type Payload struct {
	UserID string
	CID    string
	Hash   string
	AreaID string
}

// Init 初始化
func Init() {
	rCli := redis.NewClient(&redis.Options{
		Addr:     config.Cfg.RedisAddr,
		Password: config.Cfg.RedisPassword,
	})
	cli = &Client{rds: rCli}
}

// GetClient 获取redis客户端
func GetClient() *Client {
	return cli
}

// PushSchedulerInfo 插入调度器信息
func (c *Client) PushSchedulerInfo(ctx context.Context, payload *Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}

	err = c.rds.LPush(ctx, key, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l push info to redis error:%w", err)
	}

	return nil
}

// GetAllSchedulerInfos 获取所有调度器信息
func (c *Client) GetAllSchedulerInfos(ctx context.Context) ([]*Payload, error) {
	var ps []*Payload

	res, err := c.rds.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get all scheduler's info error:%w", err)
	}

	for _, v := range res {
		payload := new(Payload)
		err = json.Unmarshal([]byte(v), payload)
		if err != nil {
			log.Printf("json unmarshal scheduler's info error:%v\n", err)
			continue
		}
		ps = append(ps, payload)
	}

	return ps, nil
}

// DelSchedulerInfo 删除同步完的调度器信息
func (c *Client) DelSchedulerInfo(ctx context.Context, payload *Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}

	err = c.rds.LRem(ctx, key, 1, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l rem info to redis error:%w", err)
	}

	return nil
}
