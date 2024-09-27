package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gnasnik/titan-explorer/pkg/random"
)

var (
	apiSecret *jwt.HMACSHA
	maxAPIKey int = 20
	cryptoKey     = []byte("7c10a37c75c545daa5cff0387d46f064")
)

// JWTPayload jwt载体
type JWTPayload struct {
	// role base access controller permission
	Allow []string
	ID    string
	// TODO remove NodeID later, any role id replace as ID
	NodeID string
	// Extend is json string
	Extend string
	// The sub permission of user
	AccessControlList []userAccessControl
}

func authNew(payload *JWTPayload) (string, error) {
	tk, err := jwt.Sign(&payload, apiSecret)
	if err != nil {
		return "", err
	}

	return string(tk), nil
}

// AuthVerify verifies a JWT token and returns the permissions associated with it
func AuthVerify(token string) (*JWTPayload, error) {
	var payload JWTPayload
	if _, err := jwt.Verify([]byte(token), apiSecret, &payload); err != nil {
		return nil, fmt.Errorf("JWT Verification failed: %w", err)
	}

	// replace ID to NodeID
	if len(payload.NodeID) > 0 {
		payload.ID = payload.NodeID
	}
	return &payload, nil
}

// CreateAPIKey 创建API key
func CreateAPIKey(ctx context.Context, userID, keyName string, perms []string, buf []byte) ([]byte, string, error) {
	// check perms
	err := checkPermsIfInACL(perms)
	if err != nil {
		return nil, "", err
	}

	apiKeys := make(map[string]UserAPIKeysInfo)
	if len(buf) > 0 {
		apiKeys, err = DecodeAPIKeys(buf)
		if err != nil {
			return nil, "", err
		}
	}

	if _, ok := apiKeys[keyName]; ok {
		return nil, "", &api.ErrWeb{Code: terrors.APPKeyAlreadyExist.Int(), Message: fmt.Sprintf("the API key %s already exist", keyName)}
	}
	if len(apiKeys) >= maxAPIKey {
		return nil, "", &api.ErrWeb{Code: terrors.OutOfMaxAPIKeyLimit.Int(), Message: fmt.Sprintf("api key exceeds maximum limit %d", maxAPIKey)}
	}

	// 生成api key
	payload := JWTPayload{ID: userID, Allow: []string{"user"}, Extend: keyName, AccessControlList: perms}
	tk, err := authNew(&payload)
	if err != nil {
		return nil, "", err
	}
	apiKeys[keyName] = UserAPIKeysInfo{CreatedTime: time.Now(), APIKey: tk}
	buf, err = EncodeAPIKeys(apiKeys)
	if err != nil {
		return nil, "", err
	}

	return buf, tk, nil
}

// CreateAPIKeySecret 创建api key secret
func CreateAPIKeySecret(ctx context.Context, userID, keyName string, buf []byte) ([]byte, string, string, error) {
	var err error

	apiKeys := make(map[string]UserAPIKeySecretInfo)
	if len(buf) > 0 {
		apiKeys, err = DecodeAPIKeySecrets(buf)
		if err != nil {
			return nil, "", "", fmt.Errorf("decode UserAPIKeySecretInfo error:%w", err)
		}
	}

	if _, ok := apiKeys[keyName]; ok {
		return nil, "", "", &api.ErrWeb{Code: terrors.APPKeyAlreadyExist.Int(), Message: fmt.Sprintf("the API key %s already exist", keyName)}
	}
	if len(apiKeys) >= maxAPIKey {
		return nil, "", "", &api.ErrWeb{Code: terrors.OutOfMaxAPIKeyLimit.Int(), Message: fmt.Sprintf("api key exceeds maximum limit %d", maxAPIKey)}
	}

	// 生成api key secret
	ui := UserKeyInfo{UID: userID, Salt: random.GenerateRandomString(6)}
	uk, _ := json.Marshal(ui)
	apiKey, _ := AesEncryptCBC(uk, cryptoKey)
	apiSecret := fmt.Sprintf("ts-%s", random.GenerateRandomString(48))
	apiKeys[keyName] = UserAPIKeySecretInfo{
		APIKey:      apiKey,
		APISecret:   apiSecret,
		CreatedTime: time.Now(),
	}
	buf, err = EncodeAPIKeySecrets(apiKeys)
	if err != nil {
		return nil, "", "", fmt.Errorf("encode UserAPIKeySecretInfo error:%w", err)
	}

	return buf, apiKey, apiSecret, err
}

// AesDecryptCBCByKey 通过key解密aes密文
func AesDecryptCBCByKey(cstr string) (string, error) {
	var ukInfo UserKeyInfo

	uk, err := AesDecryptCBC(cstr, cryptoKey)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(uk, &ukInfo); err != nil {
		return "", err
	}

	return string(ukInfo.UID), nil
}

// CreateTenantKey 生成租户key
func CreateTenantKey(tenatID, name string) (string, error) {
	ui := TenantInfo{TenantID: tenatID, Salt: random.GenerateRandomString(6), Name: name}
	uk, _ := json.Marshal(ui)
	tenantKey, err := AesEncryptCBC(uk, cryptoKey)
	if err != nil {
		return "", err
	}

	return tenantKey, nil
}

// AesDecryptTenantKey 解密key并获取payload
func AesDecryptTenantKey(cstr string) (*TenantInfo, error) {
	var payload TenantInfo

	uk, err := AesDecryptCBC(cstr, cryptoKey)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(uk, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
