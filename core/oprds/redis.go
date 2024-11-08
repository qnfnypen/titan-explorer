package oprds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/go-redis/redis/v9"
)

const (
	schckey        = "sync_scheduler_list"
	areaKey        = "sync_scheduler_list_unlogin"
	preUnlogin     = "unlogin_hash"
	preNodeID      = "unsync_nodeid"
	preStorageFlow = "storage_flow"
	preDownload    = "asset_download"
	preTempFile    = "temp_file_uid"
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
	Owner  string
	AreaID string
}

// AreaIDPayload 未登陆上传文件的信息
type AreaIDPayload struct {
	CID     string
	Hash    string
	AreaIDs []string
}

// UnLoginSyncArea 未登陆同步调度器区域
type UnLoginSyncArea struct {
	List []UnloginSyncAreaDetail
}

// UnloginSyncAreaDetail 未登陆同步
type UnloginSyncAreaDetail struct {
	AreaID string
	IsSync bool
}

// Init 初始化
func Init() {
	rCli := redis.NewClient(&redis.Options{
		Addr:     config.Cfg.RedisAddr,
		Password: config.Cfg.RedisPassword,
	})
	cli = &Client{rds: rCli}
}

// GetClient 获取客户端
func GetClient() *Client {
	return cli
}

// RedisClient 获取redis客户端
func (c *Client) RedisClient() *redis.Client {
	return c.rds
}

// PushSchedulerInfo 插入调度器信息
func (c *Client) PushSchedulerInfo(ctx context.Context, payload *Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}

	err = c.rds.LPush(ctx, schckey, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l push info to redis error:%w", err)
	}

	return nil
}

// GetAllSchedulerInfos 获取所有调度器信息
func (c *Client) GetAllSchedulerInfos(ctx context.Context) ([]*Payload, error) {
	var ps []*Payload

	res, err := c.rds.LRange(ctx, schckey, 0, -1).Result()
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

	err = c.rds.LRem(ctx, schckey, 1, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l rem info to redis error:%w", err)
	}

	return nil
}

// PushAreaIDs 上传需要同步的文件区域到队列，存在则不插入
func (c *Client) PushAreaIDs(ctx context.Context, payload *AreaIDPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}

	removed, err := c.rds.LRem(ctx, areaKey, 0, string(body)).Result()
	if err != nil {
		return fmt.Errorf("l rem info to redis error:%w", err)
	}
	// 如果 remove 大于0，则说明已经存在，则直接返回
	if removed > 0 {
		return nil
	}

	err = c.rds.LPush(ctx, areaKey, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l push info to redis error:%w", err)
	}

	return nil
}

// GetAllAreaIDs 获取所有节点信息
func (c *Client) GetAllAreaIDs(ctx context.Context) ([]*AreaIDPayload, error) {
	var ps []*AreaIDPayload

	res, err := c.rds.LRange(ctx, areaKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get all scheduler's info error:%w", err)
	}

	for _, v := range res {
		payload := new(AreaIDPayload)
		err = json.Unmarshal([]byte(v), payload)
		if err != nil {
			log.Printf("json unmarshal scheduler's info error:%v\n", err)
			continue
		}
		ps = append(ps, payload)
	}

	return ps, nil
}

// DelAreaIDs 从队列删除同步完成文件区域
func (c *Client) DelAreaIDs(ctx context.Context, payload *AreaIDPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}

	err = c.rds.LRem(ctx, areaKey, 1, string(body)).Err()
	if err != nil {
		return fmt.Errorf("l rem info to redis error:%w", err)
	}

	return nil
}

// SetUnloginAssetInfo 塞入未登陆文件的hash和区域
func (c *Client) SetUnloginAssetInfo(ctx context.Context, hash string, payload *UnLoginSyncArea) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal scheduler's info error:%w", err)
	}
	key := fmt.Sprintf("%s_%s", preUnlogin, hash)

	return c.rds.SetEx(ctx, key, string(body), 24*time.Hour).Err()
}

// GetUnloginAssetAreaIDs 获取未登陆文件的区域
func (c *Client) GetUnloginAssetAreaIDs(ctx context.Context, hash string) ([]string, error) {
	var (
		payload UnLoginSyncArea
		areaIDs []string
	)

	key := fmt.Sprintf("%s_%s", preUnlogin, hash)
	value, err := c.rds.Get(ctx, key).Result()
	switch err {
	case redis.Nil:
		return nil, nil
	case nil:
		if err := json.Unmarshal([]byte(value), &payload); err != nil {
			return nil, err
		}
		for _, v := range payload.List {
			if v.IsSync {
				areaIDs = append(areaIDs, v.AreaID)
			}
		}
		return areaIDs, nil
	default:
		return nil, err
	}
}

// IncrUnSyncNodeID 增加未同步的节点次数
func (c *Client) IncrUnSyncNodeID(ctx context.Context, nodeID string) error {
	key := fmt.Sprintf("%s_%s", preNodeID, nodeID)

	if err := c.rds.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("incr key(%v) error:%w", key, err)
	}

	return nil
}

// CheckUnSyncNodeID 判断未同步的节点是否要跳过
func (c *Client) CheckUnSyncNodeID(ctx context.Context, nodeID string) (bool, error) {
	key := fmt.Sprintf("%s_%s", preNodeID, nodeID)

	num, err := c.rds.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("get value of key(%v) error:%w", key, err)
	}

	return num <= 5, nil
}

// IncrAssetHourDownload 对该时间段内文件下载量加1
func (c *Client) IncrAssetHourDownload(ctx context.Context, hash, userID string) error {
	key := fmt.Sprintf("%s_%s_%s", preDownload, userID, hash)

	return c.rds.Incr(ctx, key).Err()
}

// GetAssetHourDownload 获取该时间段内文件下载数量
func (c *Client) GetAssetHourDownload(ctx context.Context, hash, userID string) (int64, error) {
	key := fmt.Sprintf("%s_%s_%s", preDownload, userID, hash)
	v, err := c.rds.Get(ctx, key).Int64()
	switch err {
	case redis.Nil:
		return 0, nil
	case nil:
		c.rds.Del(ctx, key)
		return v, nil
	default:
		c.rds.Del(ctx, key)
		return 0, fmt.Errorf("get key error:%w", err)
	}
}

// StoreUserStorageFlowInfo 存储用户存储流量
func (c *Client) StoreUserStorageFlowInfo(ctx context.Context, uid string, value string) error {
	key := fmt.Sprintf("%s_%s", preStorageFlow, uid)

	return c.rds.Set(ctx, key, value, 2*time.Minute).Err()
}

// GetUserStorageFlowInfo 获取用户存储流量
func (c *Client) GetUserStorageFlowInfo(ctx context.Context, uid string) (string, error) {
	key := fmt.Sprintf("%s_%s", preStorageFlow, uid)

	return c.rds.Get(ctx, key).Result()
}

// SetUIDTempFileTime 设置用户临时存储文件时间
func (c *Client) SetUIDTempFileTime(ctx context.Context, uid string) error {
	key := fmt.Sprintf("%s_%s", preTempFile, uid)

	return c.rds.SetEx(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
}

// GetUIDTempFileTime 获取用户临时文件存储时间
func (c *Client) GetUIDTempFileTime(ctx context.Context, uid string) (int64, error) {
	key := fmt.Sprintf("%s_%s", preTempFile, uid)

	return c.rds.Get(ctx, key).Int64()
}
