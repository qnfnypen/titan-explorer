package filecoin

import (
	"bytes"
	"encoding/json"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/libp2p/go-libp2p/core/peer"
	"golang.org/x/xerrors"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
)

const (
	FilecoinMainnetResetTimestamp       = 1602773040
	FilecoinMainnetStartBlock           = 148888
	FilecoinMainnetEpochDurationSeconds = 30
)

type (
	// TipSet lotus struct
	TipSet struct {
		Height int64
	}

	minerInfo struct {
		PeerId           *peer.ID
		MultiAddress     [][]byte
		Owner            string
		Worker           string
		ControlAddresses []string
	}

	minerPower struct {
		MinerPower struct {
			RawBytePower    string
			QualityAdjPower string
		}

		TotalPower struct {
			RawBytePower    string
			QualityAdjPower string
		}

		HasMinPower bool
	}
)

func ChainHead(url string) (*TipSet, error) {
	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.ChainHead",
		Params:  nil,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return nil, err
	}

	var ts TipSet
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func StateMinerInfo(url string, minerId string) (*minerInfo, error) {
	params, err := json.Marshal([]interface{}{minerId, nil})
	if err != nil {
		return nil, err
	}

	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.StateMinerInfo",
		Params:  params,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return nil, err
	}

	var mi minerInfo
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &mi)
	if err != nil {
		return nil, err
	}

	return &mi, nil
}

func StateMinerPower(url string, minerId string) (*minerPower, error) {
	params, err := json.Marshal([]interface{}{minerId, nil})
	if err != nil {
		return nil, err
	}

	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.StateMinerPower",
		Params:  params,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return nil, err
	}

	var mi minerPower
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &mi)
	if err != nil {
		return nil, err
	}

	return &mi, nil
}

func StateLookupID(url string, minerId string) (string, error) {
	params, err := json.Marshal([]interface{}{minerId, nil})
	if err != nil {
		return "", err
	}

	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.StateLookupID",
		Params:  params,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return "", err
	}

	var lookupID string
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, &lookupID)
	if err != nil {
		return "", err
	}

	return lookupID, nil
}

func WalletVerify(url string, addr string, message []byte, signType byte, singData []byte) (bool, error) {
	params, err := json.Marshal([]interface{}{
		addr,
		message,
		map[string]interface{}{"Type": signType, "Data": singData}})

	if err != nil {
		return false, err
	}

	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.WalletVerify",
		Params:  params,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return false, err
	}

	var result bool
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(b, &result)
	if err != nil {
		return false, err
	}

	return result, nil
}

func WalletBalance(url string, minerId string) (string, error) {
	params, err := json.Marshal([]interface{}{minerId})
	if err != nil {
		return "", err
	}

	req := model.LotusRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin.WalletBalance",
		Params:  params,
		ID:      1,
	}

	rsp, err := requestLotus(url, req)
	if err != nil {
		return "", err
	}

	var balance string
	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, &balance)
	if err != nil {
		return "", err
	}

	return balance, nil
}

func requestLotus(url string, req model.LotusRequest) (*model.LotusResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rsp model.LotusResponse
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		return nil, err
	}

	if rsp.Error != nil {
		return nil, xerrors.New(rsp.Error.Message)
	}

	return &rsp, nil
}

func GetTimestampByHeight(height int64) int64 {
	height = height - FilecoinMainnetStartBlock
	if height < 0 {
		return 0
	}

	return FilecoinMainnetResetTimestamp + FilecoinMainnetEpochDurationSeconds*height
}

func GetReadablyBalance(balance *big.Int) float64 {
	result, err := strconv.ParseFloat(balance.String(), 64)
	if err != nil {
		return 0
	}

	return result / 1000_00000_00000_00000.0
}
