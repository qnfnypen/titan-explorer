package utils

import (
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var log = logging.Logger("utils")

const (
	TimeFormatDatetime   = "2006-01-02 15:04:05"
	TimeFormatDatetimeMc = "2006-01-02 15:04:05.000"
	TimeFormatDateOnly   = "2006-01-02"
	TimeFormatMD         = "01-02"
	TimeFormatYMDH       = "2006-01-02 15"
)

func Str2Float64(s string) float64 {
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error(err.Error())
		return 0.00
	}
	return ret
}

func StrToFloat(str string) float64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return float64(0)
	}
	return v
}

func Str2Int(s string) int {
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return ret
}

func Str2Int64(s string) int64 {
	ret, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return ret
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", num), 64)
	return num
}

func RandFloat64() float64 {
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(100)
	var randFloat float64
	if randInt%2 == 0 {
		randFloat = float64(randInt)
	} else {
		randFloat = -float64(randInt)
	}
	return randFloat
}
