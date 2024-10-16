package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/gnasnik/titan-explorer/core/storage"
	"github.com/hibiken/asynq"
)

func assetDeleteNotify(ctx context.Context, t *asynq.Task) error {

	var (
		payload opasynq.AssetUploadNotifyPayload
		err     error
	)

	err = json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		cronLog.Errorf("unable to parse message %+v", t.Payload())
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

	address, err := url.Parse(tenantInfo.DeleteNotifyUrl)
	if err != nil {
		cronLog.Errorf("invalid URL %+v", err)
		return err
	}

	var (
		secret      = pair.ApiSecret
		method      = "POST"
		url         = address.Path
		bodyData, _ = json.Marshal(payload)
	)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyData))
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
		cronLog.Errorf("received non ok http code %d with body %s", resp.StatusCode, resp.Body)
		return fmt.Errorf("received non ok http code %d with body %s", resp.StatusCode, resp.Body)
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
