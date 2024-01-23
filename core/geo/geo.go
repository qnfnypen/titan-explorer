package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	"github.com/go-redis/redis/v9"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("geo")

var (
	GEOLocationKeyPrefix = "TITAN::GEO"
)

func CacheIPLocation(ctx context.Context, location *model.Location) error {
	key := fmt.Sprintf("%s::%s", GEOLocationKeyPrefix, location.Ip)
	bytes, err := json.Marshal(location)
	if err != nil {
		return err
	}
	_, err = dao.RedisCache.Set(ctx, key, bytes, 0).Result()
	return err
}

func GetCacheLocation(ctx context.Context, ip string) (*model.Location, error) {
	key := fmt.Sprintf("%s::%s", GEOLocationKeyPrefix, ip)
	out := &model.Location{}
	bytes, err := dao.RedisCache.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetIpLocation(ctx context.Context, ip string, languages ...model.Language) (*model.Location, error) {
	//  get location from redis
	location, err := GetCacheLocation(ctx, ip)
	if err == nil && location != nil {
		return location, nil
	}

	defer func() {
		if location == nil {
			return
		}

		err = CacheIPLocation(ctx, location)
		if err != nil {
			log.Errorf("cache ip location: %v", err)
		}
	}()

	// get info from databases
	var locationdb model.Location
	err = dao.GetLocationInfoByIp(ctx, ip, &locationdb, model.LanguageEN)
	if err != nil {
		log.Errorf("get location by ip: %v", err)
		return nil, err
	}

	if locationdb != (model.Location{}) {
		return &locationdb, nil
	}

	// get location from ip data cloud api

	var lang model.Language

	if len(languages) == 0 {
		lang = model.LanguageEN
	} else {
		lang = languages[0]
	}

	for _, l := range model.SupportLanguages {
		loc, err := iptool.IPDataCloudGetLocation(ctx, config.Cfg.IpDataCloud.Url, ip, config.Cfg.IpDataCloud.Key, string(l))
		if err != nil {
			log.Errorf("ip data cloud get location: %v", err)
			continue
		}
		if err := dao.UpsertLocationInfo(ctx, loc, l); err != nil {
			continue
		}

		if lang == l {
			location = loc
		}
	}

	return location, nil
}
