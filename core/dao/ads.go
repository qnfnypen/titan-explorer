package dao

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	AdsPlatformPC  = 1
	AdsPlatformAPP = 2

	AdsTypeBanner = 1
	AdsTypeNotice = 2
)

func ListBannersCtx(ctx context.Context, platform int64, lang string) ([]*model.Ads, error) {
	var out = make([]*model.Ads, 0)

	tn := time.Now()
	query := `SELECT * FROM ads where platform = ? and lang = ? and ads_type = ? and state = ? and invalid_from <= ? and invalid_to >= ? order by weight desc, created_at desc`

	if err := DB.SelectContext(ctx, &out, query, platform, lang, AdsTypeBanner, 1, tn, tn); err != nil {
		return nil, err
	}
	return out, nil
}

func ListNoticesCtx(ctx context.Context, platform int64, lang string) ([]*model.Ads, error) {
	var n = make([]*model.Ads, 0)

	tn := time.Now()
	query := `SELECT * FROM ads where platform = ? and lang = ? and ads_type = ? and state = ? and invalid_from <= ? and invalid_to >= ? order by weight desc, created_at desc`

	if err := DB.SelectContext(ctx, &n, query, platform, lang, AdsTypeNotice, 1, tn, tn); err != nil {
		return nil, err
	}
	return n, nil
}

func AdsListPageCtx(ctx context.Context, page, size int, sb squirrel.SelectBuilder) ([]*model.Ads, int64, error) {
	var out = make([]*model.Ads, 0)
	var count int64

	if page < 1 {
		page = 1
	}

	query, args, err := sb.From("ads").Columns("*").Offset(uint64((page - 1) * size)).Limit(uint64(size)).ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := DB.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}

	query, args, err = sb.From("ads").Columns("count(*)").ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := DB.GetContext(ctx, &count, query, args...); err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

func AdsAddCtx(ctx context.Context, ads *model.Ads) error {
	query := "INSERT INTO ads (name, ads_type, redirect_url, platform, lang, `desc`, is_text, weight, state, invalid_from, invalid_to, created_at, updated_at)" +
		"VALUES (:name, :ads_type, :redirect_url, :platform, :lang, :desc, :is_text, :weight, :state, :invalid_from, :invalid_to, :created_at, :updated_at)"
	_, err := DB.NamedExecContext(ctx, query, ads)
	return err
}

func AdsDelCtx(ctx context.Context, id int64) error {
	query := `DELETE FROM ads where id = ?`
	_, err := DB.NamedExecContext(ctx, query, id)
	return err
}

func AdsUpdateCtx(ctx context.Context, ads *model.Ads) error {
	query := "UPDATE ads SET name = :name, ads_type = :ads_type, redirect_url = :redirect_url, platform = :platform, lang = :lang, " +
		"`desc` = :desc, is_text = :is_text ,weight = :weight, state = :state, hits = :hits, invalid_from = :invalid_from, invalid_to = :invalid_to, updated_at = :updated_at WHERE id = :id"
	_, err := DB.NamedExecContext(ctx, query, ads)
	return err
}

func AdsFindOne(ctx context.Context, id int64) (*model.Ads, error) {
	var out model.Ads
	query := "SELECT * FROM ads where id = ?"
	err := DB.GetContext(ctx, &out, query, id)
	return &out, err
}
