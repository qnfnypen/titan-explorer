package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/gbrlsnchs/jwt/v3"
)

var (
	apiSecret *jwt.HMACSHA
	maxAPIKey int
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
