package api

import (
	"context"
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core"
	"github.com/gnasnik/titan-explorer/core/chain"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	kub "github.com/gnasnik/titan-explorer/core/kubesphere"
	"github.com/gnasnik/titan-explorer/core/order"
	"github.com/gnasnik/titan-explorer/core/token"
	"github.com/gnasnik/titan-explorer/pkg/opcheck"
	"github.com/google/uuid"
)

var (
	mDB      *dao.Mgr
	kubMgr   *kub.Mgr
	orderMgr *order.Mgr
	tokenMgr *token.Mgr
)

// BindKeplrReq 绑定keplr请求参数
type BindKeplrReq struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Sign      string `form:"sign" json:"sign"`
}

// InitManagers 初始化platform manager配置
func InitManagers(cfg *config.Config) {
	var err error

	mDB, err = dao.NewDbMgr(cfg)
	if err != nil {
		log.Fatal("initial db err: ", err)
	}

	kubMgr, err = kub.NewKubManager(&cfg.KubesphereAPI)
	if err != nil {
		log.Fatal("initial kub err: ", err)
	}

	chainMgr, err := chain.NewChainManager(&cfg.ChainAPI)
	if err != nil {
		log.Fatal("initial chain err:", err)
	}

	orderMgr = order.NewOrderManager(mDB, kubMgr, chainMgr)
	tokenMgr = token.NewTokenManager(mDB, chainMgr)
}

func addPlatformUserInfo(ctx context.Context, su, account, userName string) error {
	_, err := mDB.GetUserMapByAccount(ctx, su)

	switch err {
	case sql.ErrNoRows:
		pwd := kubMgr.GeneratePassword(12)

		err = kubMgr.CreateUserAccount(account, pwd)
		if err != nil {
			log.Errorf("CreateUserAccount: %s", err.Error())
			return err
		}

		user := &core.User{
			Account:   account,
			Username:  userName,
			KubPwd:    pwd,
			CreatedAt: time.Now(),
		}

		err = mDB.CreateUserMap(ctx, su, user)
		if err != nil {
			return err
		}
	case nil:
	default:
		return err
	}

	return nil
}

func checkOrderParams(order *core.OrderInfoReq) int {
	if order.CPUCores > 32 || order.CPUCores < 1 {
		return errors.InvalidParams
	}

	if order.RAMSize > 64 || order.RAMSize < 1 {
		return errors.InvalidParams
	}

	if order.StorageSize > 4000 || order.StorageSize < 40 {
		return errors.InvalidParams
	}

	if order.Duration > 30*24 || order.Duration < 1 {
		return errors.InvalidParams
	}

	return 0
}

func bindKeplr(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id := claims[identityKey].(string)

	var req BindKeplrReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if checkTitanAddr(id) {
		c.JSON(http.StatusOK, respErrorCode(errors.WalletBound, c))
		return
	}

	// 判断是否已经绑定过
	user, err := mDB.GetUserMapByAccount(c.Request.Context(), id)
	switch err {
	case sql.ErrNoRows:
	case nil:
		if user.Keplr != "" {
			c.JSON(http.StatusOK, respErrorCode(errors.WalletBound, c))
			return
		}
	default:
		log.Errorf("binding error:%v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	nonce, err := getNonceFromCache(c.Request.Context(), req.Address, NonceStringTypeSignature)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	if nonce == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	match, _ := opcheck.VerifyComosSign(req.Address, nonce, req.Sign, req.PublicKey)
	if !match {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	// 添加容器平台用户信息
	err = addPlatformUserInfo(c.Request.Context(), id, req.Address, "")
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}

func getBindInfo(c *gin.Context) {
	var email string

	claims := jwt.ExtractClaims(c)
	id := claims[identityKey].(string)

	if checkIsEmail(id) {
		email = id
	}

	user, err := mDB.GetUserMapByAccount(c.Request.Context(), id)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"keplr": "",
			"email": email,
		}))
		return
	case nil:
		if user.Email == "" {
			user.Email = email
		}
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"keplr": user.Keplr,
			"email": user.Email,
		}))
		return
	default:
		log.Errorf("binding error:%v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
}

func getBindNonce(c *gin.Context) {
	username := c.Query("address")
	if username == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	nonce, err := generateNonceString(c.Request.Context(), getRedisNonceSignatureKey(username))
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"code": nonce,
	}))
}

func getUserInfoHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// id := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	id, _ := idStr.(string)

	resp, err := mDB.GetUserInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(resp))
}

func receiveTokenHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// id := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	id, _ := idStr.(string)

	code, err := tokenMgr.ReceiveTokens(id)
	if code > 0 {
		log.Errorf("receiveTokenHandler id:%s code:%d err:%v", id, code, err)
		c.JSON(http.StatusOK, respErrorCode(code, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}

func getBalanceHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// id := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	id, _ := idStr.(string)

	balance, err := tokenMgr.GetBalance(id)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.Unknown, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"balance": balance,
	}))
}

func getReceiveHistoryHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var list []*core.ReceiveHistory
	total := int64(0)

	size, _ := strconv.Atoi(c.Query("size"))
	page, _ := strconv.Atoi(c.Query("page"))

	list, total, err := mDB.LoadReceiveHistory(c, account, page, size)
	if err != nil {
		log.Errorf("getOrderHistoryHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	for _, v := range list {
		v.CreatedAt = v.CreatedAt.UTC()
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"list":  list,
		"total": total,
	}))
}

func resetKubPwdHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// id := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	id, _ := idStr.(string)

	pwd := kubMgr.GeneratePassword(12)

	err := kubMgr.ResetPassword(id, pwd)
	if err != nil {
		log.Errorf("resetKubPwdHandler id:%s err:%s", id, err.Error())
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	err = mDB.UpdateKubPwd(id, pwd)
	if err != nil {
		log.Errorf("resetKubPwdHandler UpdateKubPwd id:%s pwd:%s err:%s", id, pwd, err.Error())
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"kub_pwd": pwd,
	}))
}

func getKubURLHandler(c *gin.Context) {
	kubURL := kubMgr.GetURL()

	c.JSON(http.StatusOK, respJSON(gin.H{
		"url": kubURL,
	}))
}

func getDistributedAmountHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// id := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	id, _ := idStr.(string)

	info, err := tokenMgr.GetAmountDistributedInfo(id)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(info))
}

func getPriceHandler(c *gin.Context) {
	cpu, _ := strconv.Atoi(c.Query("cpu"))
	ram, _ := strconv.Atoi(c.Query("ram"))
	duration, _ := strconv.Atoi(c.Query("duration"))
	storage, _ := strconv.Atoi(c.Query("storage"))

	params := &core.OrderInfoReq{CPUCores: cpu, RAMSize: ram, StorageSize: storage, Duration: duration}
	if checkOrderParams(params) > 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	cost := orderMgr.CalculateTotalCost(params)

	c.JSON(http.StatusOK, respJSON(gin.H{
		"cost": cost,
	}))
}

func getRefundHandler(c *gin.Context) {
	id := c.Query("id")

	info, err := mDB.LoadOrderByID(id)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	cost := orderMgr.CalculateOrderRefund(info)

	c.JSON(http.StatusOK, respJSON(gin.H{
		"cost": cost,
	}))
}

func createOrderHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var params core.OrderInfoReq
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if checkOrderParams(&params) > 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	price := orderMgr.CalculateTotalCost(&params)

	orderID := uuid.NewString()
	err := orderMgr.Create(&params, account, orderID, price)
	if err != nil {
		log.Errorf("CreateOrder: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"id": orderID,
	}))
}

func getOrderHistoryHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var list []*core.Order
	total := int64(0)
	statsMaps := make(map[core.OrderStatus]int64)

	size, _ := strconv.Atoi(c.Query("size"))
	page, _ := strconv.Atoi(c.Query("page"))
	queryStatus := c.Query("status")
	var statuses []core.OrderStatus
	for _, s := range strings.Split(queryStatus, ",") {
		statusVal, _ := strconv.ParseInt(s, 10, 64)
		statuses = append(statuses, core.OrderStatus(statusVal))
	}

	list, total, err := mDB.LoadAccountOrdersByStatuses(account, statuses, page, size)
	if err != nil {
		log.Errorf("getOrderHistoryHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	infos, err := mDB.GetServiceNumsByStatus(account, []core.OrderStatus{
		core.OrderStatusDone, core.OrderStatusExpired, core.OrderStatusTermination})
	if err != nil {
		log.Errorf("GetServiceNumsByStatus: %v", err)
	}
	for _, v := range infos {
		statsMaps[v.Status] = v.Num
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"list":        list,
		"total":       total,
		"done":        statsMaps[core.OrderStatusDone],
		"expired":     statsMaps[core.OrderStatusExpired],
		"termination": statsMaps[core.OrderStatusTermination],
	}))
}

func terminateOrderHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var params core.OrderIDReq
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	id := params.ID

	info, err := mDB.LoadOrderByID(id)
	if err != nil {
		log.Errorf("LoadOrdersByID: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if info.Account != account {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if info.Status != core.OrderStatusDone {
		c.JSON(http.StatusOK, respErrorCode(errors.OrderStatus, c))
		return
	}

	err = orderMgr.Terminate(id, info.WorkspaceID, info.Cluster, core.OrderStatusTermination, info.Status)
	if err != nil {
		log.Errorf("TerminateOrder: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}

func renewalOrderHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var params core.OrderIDReq
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	id := params.ID

	info, err := mDB.LoadOrderByID(id)
	if err != nil {
		log.Errorf("LoadOrdersByID: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if info.Account != account {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if info.Status != core.OrderStatusDone {
		c.JSON(http.StatusOK, respErrorCode(errors.OrderStatus, c))
		return
	}

	err = orderMgr.Renewal(id)
	if err != nil {
		log.Errorf("renewalOrderHandler Renewal: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}

func upgradeOrderHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var params core.UpgradeOrderInfoReq
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	oldOrder, err := mDB.LoadOrderByID(params.ID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if oldOrder.Account != account {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if oldOrder.Status != core.OrderStatusDone {
		c.JSON(http.StatusOK, respErrorCode(errors.OrderStatus, c))
		return
	}

	orderInfo := &core.OrderInfoReq{
		CPUCores:    params.CPUCores,
		RAMSize:     params.RAMSize,
		StorageSize: params.StorageSize,
		Duration:    params.Duration,
	}

	if checkOrderParams(orderInfo) > 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	price := orderMgr.CalculateTotalCost(orderInfo)

	orderID := uuid.NewString()
	err = orderMgr.Upgrade(oldOrder, orderInfo, account, orderID, price)
	if err != nil {
		log.Errorf("Upgrade: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"id": orderID,
	}))
}

func setOrderHashHandler(c *gin.Context) {
	// claims := jwt.ExtractClaims(c)
	// account := claims[identityKey].(string)
	idStr, _ := c.Get(platformKey)
	account, _ := idStr.(string)

	var params core.OrderHashReq
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	order, err := mDB.LoadOrderByID(params.ID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	if order.Account != account {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	err = mDB.UpdateOrderHash(order.ID, params.Hash)
	if err != nil {
		log.Errorf("UpdateOrderHash err:%s", err.Error())
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}

// checkTitanAddr 校验是不是titan的钱包
func checkTitanAddr(addr string) bool {
	re := regexp.MustCompile(`^titan[a-z0-9]{32}$`)

	return re.MatchString(addr)
}

func checkIsEmail(addr string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return re.MatchString(addr)
}
