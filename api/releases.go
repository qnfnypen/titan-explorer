package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
	"time"
)

type Release struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
}

func GetReleasesHandler(c *gin.Context) {
	release, err := GetReleaseFromCache(c.Request.Context())
	if err == nil {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"release": release,
		}))
		return
	}

	//err = CacheRelease(c.Request.Context(), out)
	//if err != nil {
	//	log.Errorf("cache release: %v", err)
	//}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"release": release,
	}))
}

func UpdateReleaseInfoHandler(c *gin.Context) {
	//data, err := io.ReadAll(c.Request.Body)
	//if err != nil {
	//	c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
	//	return
	//}

	var release map[string]interface{}
	err := c.Bind(&release)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	err = CacheRelease(c.Request.Context(), release)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func CacheRelease(ctx context.Context, release interface{}) error {
	key := fmt.Sprintf("TITAN::RELEASE")

	data, err := json.Marshal(release)
	if err != nil {
		return err
	}

	_, err = dao.RedisCache.Set(ctx, key, data, 0).Result()
	if err != nil {
		log.Errorf("set release info: %v", err)
	}

	return nil
}

func GetReleaseFromCache(ctx context.Context) (map[string]interface{}, error) {
	key := fmt.Sprintf("TITAN::RELEASE")
	result, err := dao.RedisCache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err = json.Unmarshal([]byte(result), &out); err != nil {
		return nil, err
	}

	return out, nil
}

func GetAppVersionHandler(c *gin.Context) {
	platform := c.Query("platform")
	lang := model.Language(c.GetHeader("Lang"))
	if lang == "" {
		lang = model.LanguageEN
	}

	if platform == "" {
		platform = c.GetHeader("platform")
	}

	if platform == "" {
		platform = "android"
	}

	appVer, err := dao.GetLatestAppVersion(c.Request.Context(), platform, lang)
	if err != nil {
		log.Errorf("Get latest app version: %v", err)
		c.JSON(http.StatusOK, respErrorCode(http.StatusInternalServerError, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"version":     appVer.Version,
		"min_version": appVer.MinVersion,
		"description": appVer.Description,
		"url":         appVer.Url,
		"cid":         appVer.Cid,
		"size":        appVer.Size,
	}))
}

func CreateAppVersionHandler(c *gin.Context) {
	var params model.AppVersion
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	_, err := dao.GetAppVersion(c.Request.Context(), params.Version, params.Platform, model.Language(params.Lang))
	if err == nil {
		// update
		if err := dao.UpdateAppVersion(c.Request.Context(), &params); err != nil {
			log.Errorf("update app verison: %v", err)
			c.JSON(http.StatusOK, respErrorCode(http.StatusInternalServerError, c))
			return
		}

		c.JSON(http.StatusOK, respJSON(nil))
		return
	}

	if err != sql.ErrNoRows {
		log.Errorf("get app verison: %v", err)
		c.JSON(http.StatusOK, respErrorCode(http.StatusInternalServerError, c))
		return
	}

	err = dao.AddAppVersion(c.Request.Context(), &params)
	if err != nil {
		log.Errorf("add app verison: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}
