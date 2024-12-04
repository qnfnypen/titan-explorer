package job

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/gnasnik/titan-explorer/core/storage"
	"github.com/go-redis/redis"
	"github.com/hibiken/asynq"
	"github.com/jinzhu/copier"
)

type AssetUploadNotifyReq struct {
	ExtraID  string // 外部文件ID
	TenantID string // 租户ID
	UserID   string // 上传者ID

	AssetName      string
	AssetCID       string
	AssetType      string
	AssetSize      int64
	GroupID        int64
	CreatedTime    time.Time
	AssetDirectUrl string
}

func assetUploadNotify(ctx context.Context, t *asynq.Task) error {
	var (
		payload opasynq.AssetUploadNotifyPayload
		err     error
	)

	err = json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		cronLog.Errorf("unable to parse message %+v", t.Payload())
		return err
	}

	defer func(err error) {
		if err != nil {

		}
	}(err)

	var body AssetUploadNotifyReq
	if err = copier.Copy(&body, &payload); err != nil {
		cronLog.Errorf("unable to copy asset %+v", payload)
		return err
	}

	tenantInfo, err := dao.GetTenantByBuilder(ctx, squirrel.Select("*").Where("tenant_id = ?", payload.TenantID))
	if err != nil {
		cronLog.Errorf("unable to find tenant info %+v", err)
		return err
	}

	pair, err := storage.LoadTenantKeyPairFromBlob([]byte(tenantInfo.ApiKey))
	if err != nil {
		cronLog.Errorf("unable to generate secret with pair %+v", err)
		return err
	}

	address, err := url.Parse(tenantInfo.UploadNotifyUrl)
	if err != nil {
		cronLog.Errorf("invalid URL %+v", err)
		return err
	}

	var directUrl string
	if len(payload.Area) > 0 {
		schedulerClient, err := getSchedulerClient(ctx, payload.Area)
		if err == nil {

			var interval = 1 * time.Second
			var startTime = time.Now()
			var timeout = 5 * time.Second

			for {
				if time.Since(startTime) >= timeout {
					break
				}
				ret, serr := schedulerClient.ShareAssetV2(ctx, &types.ShareAssetReq{
					UserID:     payload.UserID,
					AssetCID:   payload.AssetCID,
					ExpireTime: time.Time{},
				})
				if serr != nil {
					cronLog.Errorf("ShareAssetV2 error:%#v", serr)
					cronLog.Errorf("areaIDs:%#v, userID:%s", payload.Area, payload.UserID)
				}
				if len(ret.URLs) > 0 {
					directUrl = ret.URLs[0]
					break
				}
				time.Sleep(interval)
			}

		} else {
			cronLog.Errorf("getSchedulerClient error %+v", err)
		}
	}
	body.AssetDirectUrl = directUrl

	var (
		secret = pair.ApiSecret
		// key         = pair.ApiKey
		method      = "POST"
		url         = address.Path
		bodyData, _ = json.Marshal(body)
	)

	req, err := http.NewRequest(method, tenantInfo.UploadNotifyUrl, bytes.NewBuffer(bodyData))
	if err != nil {
		cronLog.Errorf("unable to generate req %+v", err)
		return err
	}

	if err := setAuthorization(req, secret, method, url, string(bodyData)); err != nil {
		cronLog.Errorf("unable to set authorization for req %+v", err)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cronLog.Errorf("unable to send post to %s with req %v with err %+v", url, req, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent, _ := io.ReadAll(resp.Body)
		cronLog.Errorf("received non ok http code %d with body %s", resp.StatusCode, string(bodyContent))
		return fmt.Errorf("received non ok http code %d with body %s", resp.StatusCode, string(bodyContent))
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		cronLog.Errorf("unable to read response body %+v", err)
		return err
	}

	if string(respData) != "success" {
		cronLog.Errorf("unexpected resp %s", respData)
		return fmt.Errorf("unexpected resp %s", respData)
	}

	cronLog.Infof("Notified client %s, status code %d, req %+v", url, resp.StatusCode, req)

	return nil
}

func setAuthorization(req *http.Request, secret, method, path, body string) error {

	req.Header.Set("Content-Type", "application/json")

	ts := time.Now().Format(time.RFC3339)
	req.Header.Set("X-Timestamp", ts)

	c, err := rand.Int(rand.Reader, big.NewInt(1e16))
	if err != nil {
		cronLog.Errorf("unable to generate c %+v", err)
		return err
	}
	nonce := fmt.Sprintf("%d", c)
	req.Header.Set("X-Nonce", nonce)

	signature := genCallbackSignature(secret, method, path, body, ts, nonce)
	req.Header.Set("X-Signature", signature)

	cronLog.Debugf("signature: %\n", signature)
	cronLog.Debugf("apiSecret: %s\n", secret)
	cronLog.Debugf("r.Method: %s\n", method)
	cronLog.Debugf("r.URL.Path: %s\n", path)
	cronLog.Debugf("r.Body: %s\n", string(body))
	cronLog.Debugf("timestamp: %s\n", ts)
	cronLog.Debugf("nonce: %s\n", nonce)

	return nil
}

func genCallbackSignature(secret, method, path, body, timestamp, nonce string) string {
	data := method + path + body + timestamp + nonce
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

var (
	areaSchMaps              = new(sync.Map)
	DefaultAreaId            = "Asia-China-Guangdong-Shenzhen"
	SchedulerConfigKeyPrefix = "TITAN::SCHEDULERCFG"
)

// copy from api.getSchedulerClient
func getSchedulerClient(ctx context.Context, areaId string) (api.Scheduler, error) {
	v, ok := areaSchMaps.Load(areaId)
	if ok {
		return v.(api.Scheduler), nil
	}
	schedulers, err := statistics.GetSchedulerConfigs(ctx, fmt.Sprintf("%s::%s", SchedulerConfigKeyPrefix, areaId))
	if err == redis.Nil && areaId != DefaultAreaId {
		return getSchedulerClient(ctx, DefaultAreaId)
	}

	if err != nil || len(schedulers) == 0 {
		cronLog.Errorf("no scheduler found")
		return nil, errors.New("no scheduler found")
	}

	schedulerApiUrl := schedulers[0].SchedulerURL
	schedulerApiToken := schedulers[0].AccessToken
	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+schedulerApiToken)
	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
	if err != nil {
		cronLog.Errorf("create scheduler rpc client: %v", err)
		return nil, err
	}

	areaSchMaps.Store(areaId, schedulerClient)

	return schedulerClient, nil
}
