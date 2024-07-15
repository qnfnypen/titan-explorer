package geo

import (
	"context"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("geo")

func GetIpLocation(ctx context.Context, ip string, languages ...model.Language) (*model.Location, error) {
	var lang model.Language

	if len(languages) == 0 {
		lang = model.LanguageEN
	} else {
		lang = languages[0]
	}

	//  get location from redis
	location, err := dao.GetCacheLocation(ctx, ip, lang)
	if err == nil && location != nil {
		return location, nil
	}

	// get info from databases
	var locationDb model.Location
	err = dao.GetLocationInfoByIp(ctx, ip, &locationDb, lang)
	if err != nil {
		log.Errorf("get location by ip: %v", err)
		return nil, err
	}

	if locationDb != (model.Location{}) {
		return &locationDb, nil
	}

	// get location from ip data cloud api

	for _, l := range model.SupportLanguages {
		loc, err := iptool.IPDataCloudGetLocation(ctx, config.Cfg.IpDataCloud.Url, ip, config.Cfg.IpDataCloud.Key, string(l))
		if err != nil {
			log.Errorf("ip data cloud get location, ip: %s : %v", ip, err)
			continue
		}

		if err := dao.UpsertLocationInfo(ctx, loc, l); err != nil {
			log.Errorf("add location: %v", err)
			continue
		}

		err = dao.CacheIPLocation(ctx, loc, l)
		if err != nil {
			log.Errorf("cache ip location: %v", err)
		}

		if lang == l {
			location = loc
		}
	}

	return location, nil
}
