package dao

import (
	"context"
	"time"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	AdsPlatformPC  = 1
	AdsPlatformAPP = 2

	AdsTypeBanner = 1
	AdsTypeNotice = 2
)

func ListBannersCtx(ctx context.Context, platform int64, lang string) ([]*model.Ads, error) {
	var out []*model.Ads

	tn := time.Now()
	query := `SELECT * FROM ads where platform = ? and lang = ? and ads_type = ? and state = ? and invalid_from <= ? and invalid_to >= ? order by weight desc, created_at desc`

	if err := DB.SelectContext(ctx, &out, query, platform, lang, AdsTypeBanner, 1, tn, tn); err != nil {
		return nil, err
	}
	return out, nil
}

func ListNoticesCtx(ctx context.Context, platform int64, lang string) ([]*model.Ads, error) {
	var n []*model.Ads

	tn := time.Now()
	query := `SELECT * FROM ads where platform = ? and lang = ? and ads_type = ? and state = ? and invalid_from <= ? and invalid_to >= ? order by weight desc, created_at desc`

	if err := DB.SelectContext(ctx, &n, query, platform, lang, AdsTypeNotice, 1, tn, tn); err != nil {
		return nil, err
	}
	return n, nil
}
