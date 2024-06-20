package oss

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/patrickmn/go-cache"
)

type AliSTS interface {
	GetStsToken(sn string) (*StsTokenInfo, error)
}

type aliSts struct {
	Endpoint  string
	ID        string
	Secret    string
	Arn       string
	stsClient *sts.Client
	store     *cache.Cache
}

type StsTokenInfo struct {
	StatusCode      int    `json:"StatusCode"`
	AccessKeyId     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken"`
	Expiration      string `json:"Expiration"`
}

func NewSts(endpoint, id, secret, arn string) AliSTS {
	sc, err := sts.NewClientWithAccessKey(endpoint, id, secret)
	if err != nil {
		panic(err)
	}
	return &aliSts{
		Endpoint:  endpoint,
		ID:        id,
		Secret:    secret,
		Arn:       arn,
		stsClient: sc,
		store:     cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (a *aliSts) GetStsToken(sn string) (*StsTokenInfo, error) {

	req := sts.CreateAssumeRoleRequest()
	req.Scheme = "https"
	req.RoleArn = a.Arn
	req.RoleSessionName = fmt.Sprintf("sess-%s", sn)
	res, err := a.stsClient.AssumeRole(req)
	if err != nil {
		return nil, err
	}
	resp := &StsTokenInfo{}
	resp.StatusCode = res.GetOriginHttpResponse().StatusCode
	resp.AccessKeyId = res.Credentials.AccessKeyId
	resp.AccessKeySecret = res.Credentials.AccessKeySecret
	resp.Expiration = res.Credentials.Expiration
	resp.SecurityToken = res.Credentials.SecurityToken

	return resp, nil
}
