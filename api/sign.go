package api

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/filecoin"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"time"
)

func getSummaryInfo(ctx *gin.Context) {
	lang := ctx.GetHeader("Lang")
	ctx.Header("Lang", lang)

	info, err := dao.GetAllSignInfo()
	if err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))
		return
	}

	totalPower := big.NewInt(0)
	totalBalance := big.NewInt(0)

	for i := range info {
		tmpPower := big.NewInt(0)
		_, success := tmpPower.SetString(info[i].MinerPower, 10)
		if !success {
			ctx.JSON(http.StatusOK, respErrorCode(errors.ParseMinerPowerFailed, ctx))

			return
		}
		totalPower.Add(totalPower, tmpPower)

		tmpBalance := big.NewInt(0)
		_, success = tmpBalance.SetString(info[i].MinerBalance, 10)
		if !success {
			ctx.JSON(http.StatusOK, respErrorCode(errors.ParseMinerPowerFailed, ctx))

			return
		}
		totalBalance.Add(totalBalance, tmpBalance)
	}

	ctx.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
		"info": map[string]interface{}{
			"num":         len(info),
			"total_power": formatBytes(totalPower),
			"total_fil":   filecoin.GetReadablyBalance(totalBalance),
		},
	}))
}

func getCommand(ctx *gin.Context) {
	lang := ctx.GetHeader("Lang")
	ctx.Header("Lang", lang)

	var info model.SignInfo

	if err := ctx.Bind(&info); err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, ctx))

		return
	}

	infoTmp, err := dao.GetInfoByMinerID(info.MinerID)
	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))

		return
	} else if infoTmp.SignedMsg != "" {
		ctx.JSON(http.StatusOK, respErrorCode(errors.MinerIDExists, ctx))

		return
	}

	if code := checkAddress(info.MinerID, info.Address); code != 0 {
		ctx.JSON(http.StatusOK, respErrorCode(code, ctx))

		return
	}

	power, err := filecoin.StateMinerPower(config.Cfg.FilecoinRPCServerAddress, info.MinerID)
	if err != nil {
		logrus.Error(err)

		ctx.JSON(http.StatusOK, respErrorCode(errors.GetMinerPowerFailed, ctx))

		return
	}

	qualityAdjPower := big.NewInt(0)
	_, success := qualityAdjPower.SetString(power.MinerPower.QualityAdjPower, 10)
	if !success {
		ctx.JSON(http.StatusOK, respErrorCode(errors.ParseMinerPowerFailed, ctx))

		return
	}

	if qualityAdjPower.Cmp(big.NewInt(0)) == 0 {
		ctx.JSON(http.StatusOK, respErrorCode(errors.MinerPowerIsZero, ctx))

		return
	}

	minerBalance, err := filecoin.WalletBalance(config.Cfg.FilecoinRPCServerAddress, info.MinerID)
	if err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(errors.GetMinerBalanceFailed, ctx))

		return
	}

	info.SignedMsg = ""
	info.Date = time.Now().Unix()
	info.MinerPower = power.MinerPower.QualityAdjPower
	info.MinerBalance = minerBalance

	if err = dao.ReplaceSignInfo(&info); err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))

		return
	}

	msg := buildMessage(info.MinerID, info.Date)

	cmd := generateCommand(info.Address, msg)

	ctx.JSON(http.StatusOK, respJSON(JsonObject{
		"msg":     "success",
		"info":    cmd,
		"message": msg,
	}))
}

func getSignInfo(ctx *gin.Context) {
	lang := ctx.GetHeader("Lang")
	ctx.Header("Lang", lang)

	info, err := dao.GetAllSignInfo()
	if err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))
		return
	}

	for i := range info {
		qualityAdjPower := big.NewInt(0)
		_, _ = qualityAdjPower.SetString(info[i].MinerPower, 10)
		info[i].MinerPower = formatBytes(qualityAdjPower)

		info[i].MinerBalance = ""
	}

	var result string
	if len(info) != 0 {
		body, _ := json.Marshal(&info)
		result = string(body)
	}

	ctx.JSON(http.StatusOK, respJSON(JsonObject{
		"msg":  "success",
		"info": result,
	}))
}

func setSignInfo(ctx *gin.Context) {
	lang := ctx.GetHeader("Lang")
	ctx.Header("Lang", lang)

	var info model.SignInfo

	if err := ctx.Bind(&info); err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, ctx))

		return
	}

	tmp, err := dao.GetInfoByMinerID(info.MinerID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusOK, respErrorCode(errors.Unregistered, ctx))

		} else {
			ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))
		}

		return
	}

	msg := buildMessage(info.MinerID, tmp.Date)
	if code := checkSign(msg, info.SignedMsg, tmp.Address); code != 0 {
		ctx.JSON(http.StatusOK, respErrorCode(code, ctx))

		return
	}

	tmp.SignedMsg = info.SignedMsg
	info = tmp
	if err = dao.ReplaceSignInfo(&info); err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(int(terrors.DatabaseErr), ctx))

		return
	}

	ctx.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func checkAddress(minerIDStr, addrStr string) int {
	lookupID, err := filecoin.StateLookupID(config.Cfg.FilecoinRPCServerAddress, addrStr)
	if err != nil {
		logrus.Error(err)

		return errors.GetLookupIDFailed
	}

	minerInfo, err := filecoin.StateMinerInfo(config.Cfg.FilecoinRPCServerAddress, minerIDStr)
	if err != nil {
		logrus.Error(err)

		return errors.GetMinerInfoFailed
	}

	if minerInfo.Owner != lookupID && minerInfo.Worker != lookupID && !checkIsControl(minerInfo.ControlAddresses, lookupID) {
		return errors.AddressNotMatch
	}

	return 0
}

func checkIsControl(controlAddresses []string, addr string) bool {
	for i := range controlAddresses {
		if controlAddresses[i] == addr {
			return true
		}
	}

	return false
}

func buildMessage(minerID string, date int64) string {
	var message string
	message += "Signature for titan\n"
	message += minerID + "\n"
	message += time.Unix(date, 0).String()

	return message
}

func generateCommand(addr string, message string) string {
	hexMessage := hex.EncodeToString([]byte(message))

	command := fmt.Sprintf("lotus wallet sign %s %s", addr, hexMessage)

	return command
}

func checkSign(message string, hexSignedMsg string, addr string) int {
	if len(hexSignedMsg) < 2 {
		return errors.ParseSignatureFailed
	}

	signedMsg, err := hex.DecodeString(hexSignedMsg)
	if err != nil {
		return errors.ParseSignatureFailed
	}

	log.Infof("wallet verify, addr \"%s\", message \"%s\", sign type: \"%v\", sign data: \"%v\"", addr, message, signedMsg[0], signedMsg[1:])

	verify, err := filecoin.WalletVerify(config.Cfg.FilecoinRPCServerAddress, addr, []byte(message), signedMsg[0], signedMsg[1:])
	if err != nil {
		log.Errorf("sign msg failed: %s", err)

		return errors.VerifySignatureFailed
	}

	if !verify {
		return errors.SignatureError
	}

	return 0
}

func formatBytes(bytes *big.Int) string {
	base := big.NewInt(1)

	var (
		KB = big.NewInt(0).Set(base.Lsh(base, 10))
		MB = big.NewInt(0).Set(base.Lsh(base, 10))
		GB = big.NewInt(0).Set(base.Lsh(base, 10))
		TB = big.NewInt(0).Set(base.Lsh(base, 10))
		PB = big.NewInt(0).Set(base.Lsh(base, 10))
	)

	switch {
	case bytes.Cmp(PB) >= 0:
		{
			tmp := bytes.Div(bytes, TB).Int64()
			return fmt.Sprintf("%.2fPB", float64(tmp)/1024.0)
		}
	case bytes.Cmp(TB) >= 0:
		{
			tmp := bytes.Div(bytes, GB).Int64()
			return fmt.Sprintf("%.2fTB", float64(tmp)/1024.0)
		}
	case bytes.Cmp(GB) >= 0:
		{
			tmp := bytes.Div(bytes, MB).Int64()
			return fmt.Sprintf("%.2fGB", float64(tmp)/1024.0)
		}
	case bytes.Cmp(MB) >= 0:
		{
			tmp := bytes.Div(bytes, KB).Int64()
			return fmt.Sprintf("%.2fMB", float64(tmp)/1024.0)
		}
	case bytes.Cmp(KB) >= 0:
		{
			tmp := bytes.Int64()
			return fmt.Sprintf("%.2fKB", float64(tmp)/1024.0)
		}
	default:
		{
			return fmt.Sprintf("%dB", bytes.Int64())
		}
	}
}
