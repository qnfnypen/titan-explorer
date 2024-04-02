package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/mssola/user_agent"
	"net/http"
	"strings"
	"time"
)

func GetAppVersionHandler(c *gin.Context) {
	lang := model.Language(c.GetHeader("Lang"))
	if lang == "" {
		lang = model.LanguageEN
	}

	userAgent := c.GetHeader("User-Agent")

	ua := user_agent.New(userAgent)
	os := ua.OS()

	platform := "android"
	if strings.Contains(os, "iPhone") {
		platform = "ios"
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
