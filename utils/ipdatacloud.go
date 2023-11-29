package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

func IPTableCloudGetLocation(ctx context.Context, url, ip, key, lang string) (*model.Location, error) {
	reqURL := fmt.Sprintf("%s?ip=%s&key=%s&language=%s", url, ip, key, lang)
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response: %d %v", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LocationInfoRes
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	correction(&result)

	return &result.Data.Location, nil
}

type LocationInfoRes struct {
	Code int    `json:"code"`
	Data Data   `json:"data"`
	Msg  string `json:"msg"`
}

type Data struct {
	Location model.Location `json:"location"`
}

func correction(res *LocationInfoRes) {
	switch res.Data.Location.Province {
	case "Xianggang":
		res.Data.Location.Province = "HongKong"
	}

	switch res.Data.Location.City {
	case "Xianggang":
		res.Data.Location.City = "HongKong"
	}
}
