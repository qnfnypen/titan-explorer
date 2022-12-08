package utils

import (
	logging "github.com/ipfs/go-log/v2"
	"strconv"
)

var log = logging.Logger("utils")

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
