package opasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/hibiken/asynq"
)

var (
	// DefaultCli 默认客户端
	DefaultCli *Client
)

// Client asynq 客户端
type Client struct {
	cli *asynq.Client
}

// NewClient 新建客户端
func NewClient(conf asynq.RedisClientOpt) *Client {
	client := asynq.NewClient(conf)

	return &Client{cli: client}
}

// Init 初始化默认客户端
func Init() {
	DefaultCli = NewClient(asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword})
}

// EnqueueAssetGroupID 塞入用户文件的文件组id
func (c *Client) EnqueueAssetGroupID(ctx context.Context, tp AssetGroupPayload) error {
	payload, err := json.Marshal(tp)
	if err != nil {
		return fmt.Errorf("json unmarshal payload of AssetGroupID error:%w", err)
	}

	task := asynq.NewTask(TypeAssetGroupID, payload, asynq.MaxRetry(3))

	_, err = c.cli.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("could not enqueue task of AssetGroupID error:%w", err)
	}

	return nil
}

func (c *Client) EnqueueAssetUploadNotify(ctx context.Context, p AssetUploadNotifyPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("json unmarshal payload of AssetUploadNotify error:%w", err)
	}

	task := asynq.NewTask(TaskTypeAssetUploadedNotify, payload, []asynq.Option{
		asynq.MaxRetry(5),               //
		asynq.Retention(24 * time.Hour), // 任务保留一天
		asynq.Timeout(1 * time.Minute),  // 1分钟时间超时
	}...)

	_, err = c.cli.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("could not enqueue task of AssetUploadNotify error:%w", err)
	}

	return nil
}

// EnqueueDeleteAssetOperation 塞入需要删除的调度器文件
func (c *Client) EnqueueDeleteAssetOperation(ctx context.Context, tp DeleteAssetPayload) error {
	payload, err := json.Marshal(tp)
	if err != nil {
		return fmt.Errorf("json unmarshal payload of DeleteAsset error:%w", err)
	}

	task := asynq.NewTask(TypeDeleteAssetOperation, payload, asynq.MaxRetry(3))

	_, err = c.cli.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("could not enqueue task of DeleteAsset error:%w", err)
	}

	return nil
}
