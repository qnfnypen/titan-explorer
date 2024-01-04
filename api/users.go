package api

import (
	"context"
	"database/sql"
	_ "embed"
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
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/gnasnik/titan-explorer/pkg/mail"
	"github.com/gnasnik/titan-explorer/pkg/random"
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
	userInfo := &model.User{
		Username:     c.Query("username"),
		VerifyCode:   c.Query("verify_code"),
		UserEmail:    c.Query("username"),
		Referrer:     c.Query("referrer"),
		ReferralCode: random.GenerateRandomString(6),
	}

	passwd := c.Query("password")
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

	passHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.PassWordNotAllowed, c))
		return
	}
	userInfo.PassHash = string(passHash)

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
	userInfo := &model.User{
		Username:   c.Query("username"),
		VerifyCode: c.Query("verify_code"),
		UserEmail:  c.Query("username"),
	}

	passwd := c.Query("password")
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.NameNotExists, c))
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.PassWordNotAllowed, c))
		return
	}
	userInfo.PassHash = string(passHash)

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
	lang := c.GetHeader("Lang")
	userInfo.UserEmail = userInfo.Username
	key := userInfo.Username + verifyType

	vc, err := GetVerifyCode(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
		return
	}

	if vc != "" {
		c.JSON(http.StatusOK, respErrorCode(errors.GetVCFrequently, c))
		return
	}

	err = SetVerifyCode(c.Request.Context(), userInfo.Username, key, lang)
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
	timeString, _ := time.Parse(formatter.TimeFormatDatetime, timeWeb)
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

func SetVerifyCode(ctx context.Context, username, key, lang string) error {
	randNew := rand.New(rand.NewSource(time.Now().UnixNano()))
	verifyCode := fmt.Sprintf("%06d", randNew.Intn(1000000))
	bytes, err := json.Marshal(verifyCode)
	if err != nil {
		return err
	}

	err = sendEmail(username, verifyCode, lang)
	if err != nil {
		log.Errorf("send email: %v", err)
		return err
	}

	var expireTime time.Duration
	expireTime = 5 * time.Minute
	_, err = dao.Cache.Set(ctx, key, bytes, expireTime).Result()
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

//go:embed template/en/mail.html
var contentEn string

//go:embed template/cn/mail.html
var contentCn string

func sendEmail(sendTo string, vc, lang string) error {
	emailSubject := map[string]string{
		"":               "[Titan Storage] Your verification code",
		model.LanguageEN: "[Titan Storage] Your verification code",
		model.LanguageCN: "[Titan Storage] 您的验证码",
	}

	content := contentEn
	if lang == model.LanguageCN {
		content = contentCn
	}

	var verificationBtn = ""
	for _, code := range vc {
		verificationBtn += fmt.Sprintf(`<button class="button" th>%s</button>`, string(code))
	}
	content = fmt.Sprintf(content, verificationBtn)

	contentType := "text/html"
	port, err := strconv.ParseInt(config.Cfg.Email.SMTPPort, 10, 64)
	message := mail.NewEmailMessage(config.Cfg.Email.From, emailSubject[lang], contentType, content, "", []string{sendTo}, nil)
	_, err = mail.NewEmailClient(config.Cfg.Email.SMTPHost, config.Cfg.Email.Username, config.Cfg.Email.Password, int(port), message).SendMessage()
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
