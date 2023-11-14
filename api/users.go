package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
	"strconv"
	"time"
)

func GetUserInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	uuid := claims[identityKey].(string)
	user, err := dao.GetUserByUserUUID(c.Request.Context(), uuid)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
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
	if userInfo.Username == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NameExists, c))
		return
	}
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	PassHash, err := bcrypt.GenerateFromPassword([]byte(PassStr), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.PassWordNotAllowed, c))
		return
	}
	userInfo.PassHash = string(PassHash)
	verifyCode, err := GetVerifyCode(c.Request.Context(), userInfo.Username+"1")
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
		return
	}
	if verifyCode == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.VerifyCodeExpired, c))
		return
	}
	if verifyCode != userInfo.VerifyCode {
		c.JSON(http.StatusOK, respErrorCode(errors.VerifyCodeErr, c))
		return
	}
	err = dao.CreateUser(c.Request.Context(), userInfo)
	if err != nil {
		log.Errorf("create user : %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
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
		c.JSON(http.StatusOK, respErrorCode(errors.NameNotExists, c))
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	PassHash, err := bcrypt.GenerateFromPassword([]byte(PassStr), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.PassWordNotAllowed, c))
		return
	}
	userInfo.PassHash = string(PassHash)
	verifyCode, err := GetVerifyCode(c.Request.Context(), userInfo.Username+"3")
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
		return
	}
	if verifyCode == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.VerifyCodeExpired, c))
		return
	}
	if verifyCode != userInfo.VerifyCode {
		c.JSON(http.StatusOK, respErrorCode(errors.VerifyCodeErr, c))
		return
	}
	err = dao.ResetPassword(c.Request.Context(), userInfo.PassHash, userInfo.Username)
	if err != nil {
		log.Errorf("update user : %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func BeforeLogin(c *gin.Context) {
	userInfo := &model.User{}
	userInfo.Username = c.Query("username")
	UserName := userInfo.Username
	if UserName == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	_, err := dao.GetUserByUsername(c.Request.Context(), UserName)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	code, errSC := SetLoginCode(c.Request.Context(), UserName+"C")
	if errSC != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if err == nil {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"code": code,
		}))
		return
	}
	err = dao.CreateUser(c.Request.Context(), userInfo)
	if err != nil {
		log.Errorf("GetUserByUsername : %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"code": code,
	}))
}

func SetLoginCode(ctx context.Context, key string) (string, error) {
	randNew := rand.New(rand.NewSource(time.Now().UnixNano()))
	verifyCode := "TitanNetWork(" + fmt.Sprintf("%06d", randNew.Intn(1000000)) + ")"
	bytes, err := json.Marshal(verifyCode)
	if err != nil {
		return "", err
	}
	var expireTime time.Duration
	expireTime = 5 * time.Minute
	_, err = dao.Cache.Set(ctx, key, bytes, expireTime).Result()
	if err != nil {
		log.Errorf("%v:", err)
		return "", err
	}
	return verifyCode, nil
}

func GetVerifyCodeHandle(c *gin.Context) {
	userInfo := &model.User{}
	userInfo.Username = c.Query("username")
	verifyType := c.Query("type")
	userInfo.UserEmail = userInfo.Username
	err := SetVerifyCode(c.Request.Context(), userInfo.Username, userInfo.Username+verifyType)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
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
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if old != nil && old.UserID != "" && old.BindStatus == deviceInfo.BindStatus {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
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
				c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
				return
			}
		}
		if deviceInfo.BindStatus == "unbinding" {
			deviceInfo.ActiveStatus = 2
			err = schedulerClient.DeactivateNode(c.Request.Context(), deviceInfo.DeviceID, 24)
			if err != nil {
				log.Errorf("api DeactivateNode: %v", err)
				c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
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
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
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
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if old == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if old.UserID != UserID {
		c.JSON(http.StatusOK, respErrorCode(errors.UnbindingNotAllowed, c))
		return
	}

	err = dao.UpdateUserDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func DeviceUpdateHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	deviceInfo.UserID = c.Query("user_id")
	deviceInfo.DeviceName = c.Query("device_name")

	old, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceInfo.DeviceID)
	if err != nil {
		log.Errorf("get user device: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if old == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	err = dao.UpdateDeviceName(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
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

func SetPeakBandwidth(userId string) {
	peakBandwidth, e := dao.GetPeakBandwidth(context.Background(), userId)
	if e != nil {
		fmt.Printf("%v", e)
		return
	}
	var expireTime time.Duration
	expireTime = time.Hour
	_ = SetUserInfo(context.Background(), userId, peakBandwidth, expireTime)
	return
}

func SetUserInfo(ctx context.Context, key string, peakBandwidth int64, expireTime time.Duration) error {
	bytes, err := json.Marshal(peakBandwidth)
	vc := GetUserInfo(ctx, key)
	if vc != 0 {
		_, err := dao.Cache.Expire(ctx, key, expireTime).Result()
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	_, err = dao.Cache.Set(ctx, key, bytes, expireTime).Result()
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
	//var EData utils.EmailData
	//EData.Subject = "【Titan Storage】您的验证码"
	//EData.Tittle = "please check your verify code "
	//EData.SendTo = sendTo
	//EData.Content += "<p style=\"line-height:38px;margin:30px;\"> <b>亲爱的用户:</b><br>"
	//EData.Content +=
	//	"&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;您好！感谢您选择使用Titan Storage，我们是一家基" +
	//		"于Filecoin提供去中心化存储云盘服务的平台。您正在" +
	//		"进行邮箱验证，以验证您的身份或在我们的平台上进行注" +
	//		"册或登录。<br>您的验证码为：<strong>" + vc + "</strong><br>" +
	//		"&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;请在操作页面输入此验证码以完成验证。为了保证您的账" +
	//		"号安全，请勿将此验证码透露给他人。请注意，此验证码" +
	//		"在接收后的5分钟内有效。若您未在有效时间内完成验" +
	//		"证，验证码将会失效。如果验证码失效，您可以重新发起" +
	//		"邮箱验证流程获取新的验证码。如果您并未进行相关操作，" +
	//		"可能是其他用户误操作，此情况下请忽略此邮件。<br>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;感谢您" +
	//		"对Titan Storage的信任和支持，我们将一如既往地为" +
	//		"您提供高品质的服务。祝您使用愉快！<br></p>" +
	//		"<h1>Titan Storage团队</h1>"
	//err := utils.SendEmail(config.Cfg.Email, EData)
	//if err != nil {
	//	log.Errorf("sendEmailing failed:%v", err)
	//	return err
	//}
	subject := "【Titan Storage】您的验证码"
	contentType := "text/html"
	content := "<p style=\"line-height:38px;margin:30px;\"> <b>亲爱的用户:</b><br>"
	content +=
		"&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;您好！感谢您选择使用Titan Storage，我们是一家基" +
			"于Filecoin提供去中心化存储云盘服务的平台。您正在" +
			"进行邮箱验证，以验证您的身份或在我们的平台上进行注" +
			"册或登录。<br>您的验证码为：<strong>" + vc + "</strong><br>" +
			"&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;请在操作页面输入此验证码以完成验证。为了保证您的账" +
			"号安全，请勿将此验证码透露给他人。请注意，此验证码" +
			"在接收后的5分钟内有效。若您未在有效时间内完成验" +
			"证，验证码将会失效。如果验证码失效，您可以重新发起" +
			"邮箱验证流程获取新的验证码。如果您并未进行相关操作，" +
			"可能是其他用户误操作，此情况下请忽略此邮件。<br>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;感谢您" +
			"对Titan Storage的信任和支持，我们将一如既往地为" +
			"您提供高品质的服务。祝您使用愉快！<br></p>" +
			"<h1>Titan Storage团队</h1>"

	port, err := strconv.ParseInt(config.Cfg.Email.SMTPPort, 10, 64)
	message := utils.NewEmailMessage(config.Cfg.Email.Username, subject, contentType, content, "", []string{sendTo}, nil)
	_, err = utils.NewEmailClient(config.Cfg.Email.SMTPHost, config.Cfg.Email.Username, config.Cfg.Email.Password, int(port), message).SendMessage()
	if err != nil {
		return err
	}

	return nil
}

func VerifyMessage(message string, signedMessage string) (string, error) {
	// Hash the unsigned message using EIP-191
	hashedMessage := []byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(message)) + message)
	hash := crypto.Keccak256Hash(hashedMessage)
	// Get the bytes of the signed message
	decodedMessage := hexutil.MustDecode(signedMessage)
	// Handles cases where EIP-115 is not implemented (most wallets don't implement it)
	if decodedMessage[64] == 27 || decodedMessage[64] == 28 {
		decodedMessage[64] -= 27
	}
	// Recover a public key from the signed message
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), decodedMessage)
	if sigPublicKeyECDSA == nil {
		log.Errorf("Could not get a public get from the message signature")
	}
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*sigPublicKeyECDSA).String(), nil
}

func GetUserInfo(ctx context.Context, key string) int64 {
	bytes, err := dao.Cache.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return 0
	}
	if err == redis.Nil {
		return 0
	}
	var peakBandwidth int64
	err = json.Unmarshal(bytes, &peakBandwidth)
	if err != nil {
		return 0
	}
	return peakBandwidth
}

func GetUserInfoE(ctx context.Context, key string) (time.Duration, error) {
	bytes, _ := dao.Cache.TTL(ctx, key).Result()
	return bytes, nil
}
