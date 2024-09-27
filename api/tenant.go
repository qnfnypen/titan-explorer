package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

type SSOLoginReq struct {
	EntryUUID string `json:"entry_uuid"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
}

func SSOLoginHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	tenantID := claims[identityKey].(string)
	tenantName := claims[identityKey].(string)

	if tenantID == "" {
		c.JSON(401, respError(errors.InvalidAPPKey, fmt.Errorf("missing app_key in request")))
		return
	}

	var req SSOLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, respError(errors.InvalidParams, fmt.Errorf("invalid sso login payload")))
		return
	}

	if req.EntryUUID == "" {
		c.JSON(200, respError(errors.InvalidParams, fmt.Errorf("invalid entry uuid")))
		return
	}

	if req.Username == "" {
		c.JSON(200, respError(errors.InvalidParams, fmt.Errorf("invalid entry username")))
		return
	}

	tenant, err := dao.GetTenantByBuilder(c.Request.Context(), squirrel.Select("*").Where("tenant = ?", tenantID))
	if err != nil {
		log.Errorf("[TENANT][SSO] query tenant error: %s", err.Error())
		c.JSON(200, respErrorCode(errors.InternalServer, c))
		return
	}

	user, err := dao.GetUserByBuilder(c.Request.Context(), squirrel.Select("*").Where(squirrel.Eq{"uuid": req.EntryUUID, "tenant_id": tenantID}))
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("[TENANT][SSO] query user error: %s", err.Error())
		c.JSON(200, respErrorCode(errors.InternalServer, c))
		return
	}

	if err == sql.ErrNoRows {
		// insert user
		user = &model.User{Uuid: req.EntryUUID, TenantID: tenantID, Avatar: req.Avatar, UserEmail: req.Email, Username: fmt.Sprintf("%s/%s", tenantName, req.Username), CreatedAt: time.Now()}
		if err := dao.CreateUser(c.Request.Context(), user); err != nil {
			log.Errorf("[TENANT][SSO] create user error: %s", err.Error())
			c.JSON(200, respErrorCode(errors.InternalServer, c))
		}
	}

	payloadProto := &model.User{TenantID: tenant.TenantID, Uuid: user.Uuid, Username: user.Username, Role: user.Role}
	token, expireTime, err := authMiddleware.TokenGenerator(payloadProto)
	if err != nil {
		log.Errorf("[TENANT][SSO] error while generating token: %s", err.Error())
		c.JSON(200, respErrorCode(errors.InternalServer, c))
	}

	c.JSON(200, respJSON(JsonObject{
		"token": token,
		"exp":   expireTime.Unix(),
	}))
}

func SubUserSyncHandler(c *gin.Context) {

}

func SubUserDeleteHandler(c *gin.Context) {

}

func SubUserRefreshTokenHandler(c *gin.Context) {

	// refreshTokenFunc := func(c *gin.Context, code int, token string, expire time.Time) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code":   http.StatusOK,
	// 		"token":  token,
	// 		"expire": expire.Format(time.RFC3339),
	// 	})
	// }

}
