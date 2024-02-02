package api

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/random"
	"net/http"
	"strings"
	"time"
)

// AuthAPIKeyMiddlewareFunc makes GinJWTMiddleware implement the Middleware interface.
func AuthAPIKeyMiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, respErrorCode(errors.NoBearerToken, c))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			c.JSON(http.StatusUnauthorized, respErrorCode(errors.NoBearerToken, c))
			c.Abort()
			return
		}

		secret, err := dao.GetSecretKey(c.Request.Context(), token)
		if err != nil {
			log.Errorf("get secret key: %v", err)
			c.JSON(http.StatusUnauthorized, respErrorCode(errors.InvalidAPPKey, c))
			c.Abort()
			return
		}

		if secret.Status == 1 {
			c.JSON(http.StatusUnauthorized, respErrorCode(errors.InvalidAPPKey, c))
			c.Abort()
			return
		}

		c.Set("user_id", secret.UserID)
	}
}

func CreateNewSecretKeyHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	//perms := c.Query("perm")

	appKey := random.GenerateRandomString(18)
	appSecret := fmt.Sprintf("ts-%s", random.GenerateRandomString(48))

	err := dao.AddUserSecret(c.Request.Context(), &model.UserSecret{
		UserID:    username,
		AppKey:    appKey,
		AppSecret: appSecret,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Errorf("Add User secret: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"app_key":    appKey,
		"app_secret": appSecret,
	}))
}
