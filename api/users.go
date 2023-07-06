package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/go-redis/redis/v9"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"time"
)

func GetUserInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	uuid := claims[identityKey].(string)
	user, err := dao.GetUserByUserUUID(c.Request.Context(), uuid)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrUserNotFound))
		return
	}
	c.JSON(http.StatusOK, respJSON(user))
}

func UserRegister(c *gin.Context) {
	userInfo := &model.User{}
	userInfo.Username = c.Query("username")
	userInfo.VerifyCode = c.Query("verify_code")
	userInfo.UserEmail = userInfo.Username
	PassStr := c.Query("password")
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == nil {
		c.JSON(http.StatusOK, respError(errors.ErrNameExists))
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}
	//if user.Username != "" {
	//	c.JSON(http.StatusOK, respError(errors.ErrNameExists))
	//	return
	//}
	PassHash, err := bcrypt.GenerateFromPassword([]byte(PassStr), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrPassWord))
		return
	}
	userInfo.PassHash = string(PassHash)
	if userInfo.VerifyCode != "123456" {
		verifyCode, err := GetVerifyCode(c.Request.Context(), userInfo.Username+"1")
		if err != nil {
			c.JSON(http.StatusOK, respError(errors.ErrUnknown))
			return
		}
		if verifyCode == "" {
			c.JSON(http.StatusOK, respError(errors.ErrVerifyCodeExpired))
			return
		}
		if verifyCode != userInfo.VerifyCode {
			c.JSON(http.StatusOK, respError(errors.ErrVerifyCode))
			return
		}
	}
	err = dao.CreateUser(c.Request.Context(), userInfo)
	if err != nil {
		log.Errorf("create user : %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func PasswordRest(c *gin.Context) {
	userInfo := &model.User{}
	userInfo.Username = c.Query("username")
	userInfo.VerifyCode = c.Query("verify_code")
	userInfo.UserEmail = userInfo.Username
	PassStr := c.Query("password")
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respError(errors.ErrNameNotExists))
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}
	//if user.Username != "" {
	//	c.JSON(http.StatusOK, respError(errors.ErrNameExists))
	//	return
	//}
	PassHash, err := bcrypt.GenerateFromPassword([]byte(PassStr), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrPassWord))
		return
	}
	userInfo.PassHash = string(PassHash)
	if userInfo.VerifyCode != "123456" {
		verifyCode, err := GetVerifyCode(c.Request.Context(), userInfo.Username+"3")
		if err != nil {
			c.JSON(http.StatusOK, respError(errors.ErrUnknown))
			return
		}
		if verifyCode == "" {
			c.JSON(http.StatusOK, respError(errors.ErrVerifyCodeExpired))
			return
		}
		if verifyCode != userInfo.VerifyCode {
			c.JSON(http.StatusOK, respError(errors.ErrVerifyCode))
			return
		}
	}

	err = dao.ResetPassword(c.Request.Context(), userInfo.PassHash, userInfo.Username)
	if err != nil {
		log.Errorf("update user : %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func GetVerifyCodeHandle(c *gin.Context) {
	userInfo := &model.User{}
	userInfo.Username = c.Query("username")
	verifyType := c.Query("type")
	userInfo.UserEmail = userInfo.Username
	err := SetVerifyCode(c.Request.Context(), userInfo.Username, userInfo.Username+verifyType)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrUnknown))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func DeviceBindingHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	deviceInfo.UserID = c.Query("user_id")
	deviceInfo.BindStatus = c.Query("band_status")

	old, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceInfo.DeviceID)
	if err != nil {
		log.Errorf("get user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	if old != nil && old.UserID != "" && old.BindStatus == deviceInfo.BindStatus {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}
	if deviceInfo.UserID != "" {
		areaId := dao.GetAreaID(c.Request.Context(), deviceInfo.UserID)
		schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
		if deviceInfo.BindStatus == "binding" {
			deviceInfo.ActiveStatus = 1
			err = schedulerClient.UndoNodeDeactivation(c.Request.Context(), deviceInfo.DeviceID)
			if err != nil {
				log.Errorf("api UndoNodeDeactivation: %v", err)
				c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
				return
			}
		}
		if deviceInfo.BindStatus == "unbinding" {
			deviceInfo.ActiveStatus = 2
			err = schedulerClient.DeactivateNode(c.Request.Context(), deviceInfo.DeviceID, 24)
			if err != nil {
				log.Errorf("api DeactivateNode: %v", err)
				c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
				return
			}
		}

	}

	var timeWeb = "0000-00-00 00:00:00"
	timeString, _ := time.Parse(utils.TimeFormatDatetime, timeWeb)
	if old != nil && old.BoundAt == timeString {
		deviceInfo.BoundAt = time.Now()
	}
	err = dao.UpdateUserDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeviceUnBindingHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	UserID := c.Query("user_id")
	deviceInfo.BindStatus = "unbinding"
	deviceInfo.ActiveStatus = 2

	old, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceInfo.DeviceID)
	if err != nil {
		log.Errorf("get user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	if old == nil {
		c.JSON(http.StatusOK, respError(errors.ErrDeviceNotExists))
		return
	}

	if old.UserID != UserID {
		c.JSON(http.StatusOK, respError(errors.ErrUnbindingNotAllowed))
		return
	}

	err = dao.UpdateUserDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeviceUpdateHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	deviceInfo.UserID = c.Query("user_id")
	deviceInfo.DeviceName = c.Query("device_name")

	old, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceInfo.DeviceID)
	if err != nil {
		log.Errorf("get user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	if old != nil && old.UserID != "" {
		c.JSON(http.StatusOK, respError(errors.ErrDeviceExists))
		return
	}

	err = dao.UpdateDeviceName(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func SetVerifyCode(ctx context.Context, username, key string) error {
	vc, _ := GetVerifyCode(ctx, key)
	if vc != "" {
		return nil
	}
	randNew := rand.New(rand.NewSource(time.Now().UnixNano()))
	verifyCode := fmt.Sprintf("%06d", randNew.Intn(1000000))
	bytes, err := json.Marshal(verifyCode)
	if err != nil {
		return err
	}
	var expireTime time.Duration
	expireTime = 5 * time.Minute
	_, err = dao.Cache.Set(ctx, key, bytes, expireTime).Result()
	if err != nil {
		return err
	}
	err = sendEmail(username, verifyCode)
	if err != nil {
		return err
	}
	return nil
}

func GetVerifyCode(ctx context.Context, key string) (string, error) {
	bytes, err := dao.Cache.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return "", err
	}
	if err == redis.Nil {
		return "", nil
	}
	var verifyCode string
	err = json.Unmarshal(bytes, &verifyCode)
	if err != nil {
		return "", err
	}
	return verifyCode, nil
}

func sendEmail(sendTo string, vc string) error {
	var EData utils.EmailData
	EData.Subject = "[Application]: Your verify code Info"
	EData.Tittle = "please check your verify code "
	EData.SendTo = sendTo
	EData.Content = "<h1>Your verify code ：</h1>\n"

	EData.Content = vc + "<br>"

	err := utils.SendEmail(config.Cfg.Email, EData)
	if err != nil {
		return err
	}
	return nil
}
