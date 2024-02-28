package api

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/filecoin"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func getSignInfo(ctx *gin.Context) {
	info, err := dao.GetAllSignInfo()
	if err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(errors.InternalServer, ctx))
		return
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
	var info model.SignInfo

	if err := ctx.Bind(&info); err != nil {
		ctx.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, ctx))

		return
	}

	if code := ValidateBasic(&info); code != 0 {
		ctx.JSON(http.StatusOK, respErrorCode(code, ctx))

		return
	}

	if err := dao.InsertSignInfo(&info); err != nil {
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				ctx.JSON(http.StatusOK, respErrorCode(errors.MinerIDExists, ctx))
			} else {
				ctx.JSON(http.StatusOK, respErrorCode(errors.InternalServer, ctx))
			}
		}

		return
	}

	ctx.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func ValidateBasic(s *model.SignInfo) int {
	if code := checkAddress(s.MinerID, s.Address); code != 0 {
		return code
	}

	message := buildMessage(s.MinerID, s.Date)

	return checkSign(message, s.SignedMsg, s.Address)
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

	if minerInfo.Owner != lookupID && minerInfo.Worker != lookupID {
		return errors.AddressNotMatch
	}

	power, err := filecoin.StateMinerPower(config.Cfg.FilecoinRPCServerAddress, minerIDStr)
	if err != nil {
		logrus.Error(err)

		return errors.GetMinerPowerFailed
	} else if power.TotalPower.RawBytePower == "" {
		return errors.MinerPowerIsZero
	}

	return 0
}

func buildMessage(minerID string, date int64) string {
	var message string
	message += "Signature for titan\n"
	message += minerID + "\n"
	message += time.Unix(date, 0).String()

	return message
}

func checkSign(message string, hexSignedMsg string, addr string) int {
	signedMsg, err := hex.DecodeString(hexSignedMsg)
	if err != nil {
		return errors.ParseSignatureFailed
	}

	verify, err := filecoin.WalletVerify(config.Cfg.FilecoinRPCServerAddress, addr, message, signedMsg[0], signedMsg[1:])
	if err != nil {
		return errors.VerifySignatureFailed
	}

	if !verify {
		return errors.SignatureError
	}

	return 0
}
