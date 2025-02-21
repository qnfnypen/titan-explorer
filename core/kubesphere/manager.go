package kub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	netURL "net/url"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/oprds"

	logging "github.com/ipfs/go-log/v2"
	"github.com/jmoiron/sqlx"
)

var log = logging.Logger("kubesphere")

const (
	timeInterval   = 10 * time.Minute
	kuberToken     = "kuber_token"
	kuberTokenLock = "kuber_token_lock"
)

// Mgr manages Kubernetes resources.
type Mgr struct {
	db            *sqlx.DB
	kubURL        string
	adminAccount  string
	adminPassword string
	curCluster    string

	token string
}

// NewKubManager creates a new instance of Mgr with the provided configuration.
func NewKubManager(cfg *config.KubesphereAPIConfig) (*Mgr, error) {
	m := &Mgr{}

	m.kubURL = cfg.URL
	m.adminAccount = cfg.AdminAccount
	m.adminPassword = cfg.AdminPassword
	m.curCluster = cfg.Cluster

	go m.startTimer()

	for {
		token, err := m.getToken()
		if err != nil {
			return nil, err
		}
		time.Sleep(2 * time.Second)
		if token != "" {
			m.token = token
			return m, nil
		}
	}
}

func (m *Mgr) startTimer() {
	ticker := time.NewTicker(timeInterval)
	defer ticker.Stop()

	for {
		<-ticker.C

		token, err := m.getToken()
		if err != nil {
			log.Errorf("getToken err: %s", err.Error())
			continue
		}
		m.token = token
	}
}

// func(m *Mgr) test() {
// 	userAccount := "cosmos12345670000002"
// 	// err := CreateUserAccount(userName)
// 	// if err != nil {
// 	// 	log.Errorf("CreateUserAccount: %s", err.Error())
// 	// 	return
// 	// }

// 	workspaceID := "order00000005"
// 	err := createUserSpace(workspaceID, userAccount, curCluster)
// 	if err != nil {
// 		log.Errorf("CreateUserSpace: %s", err.Error())
// 		return
// 	}

// 	time.Sleep(1 * time.Second)
// 	err = changeWorkspaceMembers(workspaceID, userAccount)
// 	if err != nil {
// 		log.Errorf("changeWorkspaceMembers: %s", err.Error())
// 	}
// 	err = createUserResourceQuotas(workspaceID, curCluster, 2, 2, 10)
// 	if err != nil {
// 		log.Errorf("CreateUserResourceQuotas: %s", err.Error())
// 	}

// 	err = DeleteUserSpace("order00000004", curCluster)
// 	if err != nil {
// 		log.Errorf("DeleteUserSpace: %s", err.Error())
// 	}
// }

func (m *Mgr) getToken() (string, error) {
	// 判断 token 是否存在，不存在再重新获取
	token, _ := oprds.GetClient().RedisClient().Get(context.Background(), kuberToken).Result()
	ttl, _ := oprds.GetClient().RedisClient().TTL(context.Background(), kuberToken).Result()
	if token != "" && ttl > 20*time.Minute {
		return token, nil
	}
	res, err := oprds.GetClient().RedisClient().SetNX(context.Background(), kuberTokenLock, "1", 10*time.Second).Result()
	if err != nil || !res {
		// return nil, errors.CustomError("Please try again later")
		return "", nil
	}
	data := netURL.Values{}
	data.Set("grant_type", "password")
	data.Set("username", m.adminAccount)
	data.Set("password", m.adminPassword)
	data.Set("client_id", "kubesphere")
	data.Set("client_secret", "kubesphere")

	client := &http.Client{}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", m.kubURL, "/oauth/token"), nil)
	if err != nil {
		log.Errorf("Error creating request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(data.Encode()))

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error sending request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response: %v", err)
		return "", err
	}

	log.Infof("Response Status: %s\n", resp.Status)
	log.Infof("Response Body: %s\n", string(body))

	var tokenResp tokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		log.Errorf("Error unmarshalling JSON: %v", err)
		return "", err
	}

	oprds.GetClient().RedisClient().Set(context.Background(), kuberToken, tokenResp.AccessToken, 100*time.Minute)

	return tokenResp.AccessToken, nil
}

func (m *Mgr) doRequest(method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", m.kubURL, path)

	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+m.token)
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
