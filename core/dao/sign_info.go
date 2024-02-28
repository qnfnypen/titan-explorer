package dao

import (
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func InsertSignInfo(info *model.SignInfo) error {
	_, err := DB.NamedExec("INSERT INTO sign_info (miner_id, address, date, signed_msg) VALUES (:miner_id, :address, :date, :signed_msg)", info)

	return err
}

func GetAllSignInfo() ([]model.SignInfo, error) {
	var infos []model.SignInfo

	err := DB.Select(&infos, "select * from sign_info")
	return infos, err
}
