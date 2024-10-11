package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gnasnik/titan-explorer/pkg/random"
)

// CreateTenantKey 生成租户key
func CreateTenantKey(tenatID, name string) ([]byte, string, string, error) {
	tenantInfo := TenantInfo{TenantID: tenatID, Salt: random.GenerateRandomString(6), Name: name}
	tk, _ := json.Marshal(tenantInfo)

	apiKey, err := AesEncryptCBC(tk, cryptoKey)
	if err != nil {
		return nil, "", "", err
	}

	apiSecret := fmt.Sprintf("tn-%s", random.GenerateRandomString(48))

	buf, err := GenTenantKeySecretBlob(apiKey, apiSecret)
	if err != nil {
		return nil, "", "", err
	}

	return buf, apiKey, apiSecret, nil
}

type TenantKeySecretPair struct {
	ApiKey      string
	ApiSecret   string
	CreatedTime time.Time
}

func GenTenantKeySecretBlob(apiKey, apiSecret string) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	data := &TenantKeySecretPair{
		ApiKey:      apiKey,
		ApiSecret:   apiSecret,
		CreatedTime: time.Now(),
	}

	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func LoadTenantKeyPairFromBlob(buf []byte) (*TenantKeySecretPair, error) {
	var buffer bytes.Buffer
	buffer.Write(buf)

	dec := gob.NewDecoder(&buffer)

	var data *TenantKeySecretPair
	err := dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
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
