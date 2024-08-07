package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func ListAdsHandler(c *gin.Context) {
	size, _ := strconv.Atoi(c.Query("size"))
	page, _ := strconv.Atoi(c.Query("page"))

	state, _ := strconv.Atoi(c.Query("state"))
	adsType, _ := strconv.Atoi(c.Query("ads_type"))
	platform, _ := strconv.Atoi(c.Query("platform"))
	lang := c.Query("lang")

	sb := squirrel.Select()
	if state > 0 {
		sb = sb.Where("state = ?", state)
	}
	if adsType > 0 {
		sb = sb.Where("ads_type = ?", adsType)
	}
	if lang != "" {
		sb = sb.Where("lang = ?", lang)
	}
	if platform > 0 {
		sb = sb.Where("pplatform = ?", platform)
	}

	list, total, err := dao.AdsListPageCtx(c.Request.Context(), page, size, sb)
	if err != nil {
		log.Errorf("ListAdsHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.AdsFetchFailed, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func AddAdsHandler(c *gin.Context) {
	var ads model.Ads
	if err := c.BindJSON(&ads); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if ads.AdsType != dao.AdsTypeBanner && ads.AdsType != dao.AdsTypeNotice {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if ads.Lang != "cn" && ads.Lang != "en" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if ads.InvalidFrom.Unix() > ads.InvalidTo.Unix() {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	ads.CreatedAt = time.Now()
	ads.UpdatedAt = time.Now()

	if err := dao.AdsAddCtx(c.Request.Context(), &ads); err != nil {
		log.Errorf("AddAdsHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeleteAdsHandler(c *gin.Context) {
	var ads model.Ads
	if err := c.BindJSON(&ads); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	if err := dao.AdsDelCtx(c.Request.Context(), ads.ID); err != nil {
		log.Errorf("DeleteAdsHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func UpdateAdsHandler(c *gin.Context) {
	var ads model.Ads
	if err := c.BindJSON(&ads); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	ads.UpdatedAt = time.Now()
	if err := dao.AdsUpdateCtx(c.Request.Context(), &ads); err != nil {
		log.Errorf("UpdateAdsHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}
