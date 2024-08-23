package dao

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableAreaMap = "area_map"
)

// GetAreaCnByAreaEn 通过英文的区域名称获取中文的区域名称
func GetAreaCnByAreaEn(ctx context.Context, areaEn []string) ([]string, error) {
	var areaCn []string

	query, args, err := squirrel.Select("area_cn").From(tableAreaMap).Where(squirrel.Eq{
		"area_en": areaEn,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get areas error:%w", err)
	}

	err = DB.SelectContext(ctx, &areaCn, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get areas error:%w", err)
	}

	return areaCn, nil
}

// GetAreaEnByAreaCn 通过中文的区域名称获取英文的区域名称
func GetAreaEnByAreaCn(ctx context.Context, areaCn []string) ([]string, error) {
	var areaEn []string

	query, args, err := squirrel.Select("area_en").From(tableAreaMap).Where(squirrel.Eq{
		"area_cn": areaCn,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get areas error:%w", err)
	}

	err = DB.SelectContext(ctx, &areaEn, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get areas error:%w", err)
	}

	return areaEn, nil
}

// GetAreaMapByEn 通过英文获取节点区域列表
func GetAreaMapByEn(ctx context.Context, areaEn []string) ([]model.AreaMap, error) {
	var list []model.AreaMap

	query, args, err := squirrel.Select("*").From(tableAreaMap).Where(squirrel.Eq{
		"area_en": areaEn,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate get areas error:%w", err)
	}

	err = DB.SelectContext(ctx, &list, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get areas error:%w", err)
	}

	return list, nil
}
