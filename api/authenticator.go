package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
	username := c.Query("user_id")
	appKey := randStr(18)
	appSecret := fmt.Sprintf("ts-%s", randStr(48))

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

// n is the length of random string we want to generate
func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
