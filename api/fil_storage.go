package api

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/multiformats/go-multiaddr"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	FilcoinMainnetResetTimestamp       = 1602773040
	FilcoinMainnetStartBlock           = 148888
	FilcoinMainnetEpochDurationSeconds = 30
)

func CreateFilStorageHandler(c *gin.Context) {
	var params []*model.FilStorage
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	for i := 0; i < len(params); i++ {
		params[i].StartTime = time.Unix(getTimestampByHeight(params[i].StartHeight), 0)
		params[i].EndTime = time.Unix(getTimestampByHeight(params[i].EndHeight), 0)
		params[i].CreatedAt = time.Now()
		params[i].UpdatedAt = time.Now()
		if params[i].FIndex == 0 {
			params[i].FIndex = int64(i)
		}

		sp, err := dao.GetStorageProvider(c.Request.Context(), params[i].Provider)
		if err == nil && sp != nil {
			continue
		}

		err = SaveProviderLocation(params[i].Provider)
		if err != nil {
			log.Errorf("SaveProviderLocation: %v", err)
		}
	}

	if err := dao.AddFilStorages(c.Request.Context(), params); err != nil {
		log.Errorf("add fil storage: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetFilStorageListHandler(c *gin.Context) {
	cid := c.Query("cid")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	asset, err := dao.GetAssetByCID(c.Request.Context(), cid)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"list":  nil,
			"total": 0,
		}))
		return
	}

	if err != nil {
		log.Errorf("GetAssetByCID: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	option := dao.QueryOption{
		Page:     page,
		PageSize: pageSize,
	}

	list, total, err := dao.ListFilStorages(c.Request.Context(), asset.Path, option)
	if err != nil {
		log.Errorf("ListFilStorages: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func getTimestampByHeight(height int64) int64 {
	height = height - FilcoinMainnetStartBlock
	if height < 0 {
		return 0
	}

	return FilcoinMainnetResetTimestamp + FilcoinMainnetEpochDurationSeconds*height
}

func SaveProviderLocation(providerId string) error {
	url := "http://api.node.glif.io/rpc/v0"
	minerInfo, err := StateMinerInfo(url, providerId)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if len(minerInfo.Multiaddrs) == 0 {
		return dao.AddStorageProvider(ctx, &model.StorageProvider{
			ProviderID:  providerId,
			IP:          "",
			Retrievable: false,
			Location:    "",
			CreatedAt:   time.Now(),
		})
	}

	mad, err := multiaddr.NewMultiaddrBytes(minerInfo.Multiaddrs[0])
	if err != nil {
		return err
	}

	ip := strings.Split(mad.String(), "/")[2]
	loc, err := utils.IPTableCloudGetLocation(ctx, config.Cfg.IpUrl, ip, config.Cfg.IpKey, model.LanguageEN)
	if err != nil {
		return err
	}

	return dao.AddStorageProvider(ctx, &model.StorageProvider{
		ProviderID:  providerId,
		IP:          ip,
		Retrievable: true,
		Location:    loc.CountryCode,
		CreatedAt:   time.Now(),
	})
}
