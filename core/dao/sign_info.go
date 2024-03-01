package dao

import (
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func GetInfoByMinerID(minerID string) (model.SignInfo, error) {
	var info model.SignInfo
	err := DB.Get(&info, "SELECT * FROM sign_info WHERE miner_id = ?", minerID)

	return info, err
}

func ReplaceSignInfo(info *model.SignInfo) error {
	_, err := DB.NamedExec("REPLACE INTO sign_info (miner_id, address, date, signed_msg) VALUES (:miner_id, :address, :date, :signed_msg)", info)

	return err
}

func GetAllSignInfo() ([]model.SignInfo, error) {
	var infos []model.SignInfo

	err := DB.Select(&infos, "select * from sign_info")
	return infos, err
}
