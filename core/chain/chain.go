package chain

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gnasnik/titan-explorer/config"

	chaintypes "github.com/Titannet-dao/titan-chain/x/wasm/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("chain")

// Mgr manages the chain operations.
type Mgr struct {
	txClient *cosmosclient.Client
	qClient  chaintypes.QueryClient

	prefix        string
	rpc           string
	tokenContract string
	serviceName   string
	keyringDir    string
	faucetGas     string
	orderContract string
}

// NewChainManager creates a new instance of Mgr with the provided configuration.
func NewChainManager(cfg *config.ChainAPIConfig) (*Mgr, error) {
	m := &Mgr{}

	m.prefix = cfg.AddressPrefix
	m.rpc = cfg.RPC
	m.tokenContract = cfg.TokenContractAddress
	m.serviceName = cfg.ServiceName
	m.keyringDir = cfg.KeyringDir
	m.faucetGas = cfg.FaucetGas
	m.orderContract = cfg.OrderContractAddress

	tc, err := cosmosclient.New(context.Background(),
		cosmosclient.WithAddressPrefix(m.prefix),
		cosmosclient.WithNodeAddress(m.rpc),
		cosmosclient.WithGas("600000"),
		cosmosclient.WithGasPrices("0.0025uttnt"),
		cosmosclient.WithKeyringServiceName(m.serviceName),
		cosmosclient.WithKeyringDir(m.keyringDir),
	)
	if err != nil {
		return nil, err
	}

	qc := chaintypes.NewQueryClient(tc.Context())

	m.qClient = qc
	m.txClient = &tc

	return m, nil
}

// func (m *Mgr) testSend() {
// balance, err := balance("titan17ljevhtqu4vx6y7k743jyca0w8gyfu2466e8x3")
// log.Infof("balance coin:%s", balance)
// coin := order.CalculateTotalCost(&core.OrderReq{CPUCores: 4, RAMSize: 4, StorageSize: 50, Duration: 12})
// log.Infof("CalculateTotalCost coin:%d", coin)
// err = sendOrder("order_123456", 4, 4, 50, 12, fmt.Sprintf("%d", coin))
// log.Infof("sendOrder err:%v", err)
// 	// outputs := []banktypes.Output{
// 	// 	{
// 	// 		Address: "titan17ljevhtqu4vx6y7k743jyca0w8gyfu2466e8x3",
// 	// 		Coins:   cosmostypes.NewCoins(cosmostypes.Coin{Denom: "", Amount: math.NewInt(1000000)}),
// 	// 	},
// 	// }
// 	// err := SendMsgs(outputs)
// 	// if err != nil {
// 	// 	log.Errorf("testSend err:%s", err.Error())
// 	// }
// 	toAddress := "titan17ljevhtqu4vx6y7k743jyca0w8gyfu2466e8x3"

// 	err := faucetSend(toAddress)
// 	if err != nil {
// 		log.Errorf("SendMsg err:%s", err.Error())
// 	}
// }

func (m *Mgr) getAccount() *cosmosaccount.Account {
	acc, err := m.txClient.Account(m.serviceName)
	if err != nil {
		return nil
	}

	return &acc
}

// ReceiveTokens transfers tokens to the specified address.
func (m *Mgr) ReceiveTokens(toAddress string, faucetToken string) (string, error) {
	a := m.getAccount()
	if a == nil {
		return "", errors.New("no account found")
	}

	faucetAddr, err := a.Address(m.prefix)
	if err != nil {
		return "", err
	}

	// 合约代币
	tokenBody := map[string]interface{}{
		"transfer": map[string]interface{}{
			"recipient": toAddress,
			"amount":    faucetToken,
		},
	}

	tokenJSONBody, err := json.Marshal(tokenBody)
	if err != nil {
		return "", err
	}

	tokenReq := &chaintypes.MsgExecuteContract{Sender: faucetAddr, Contract: m.tokenContract, Msg: tokenJSONBody}

	// 主币, 作为gas
	gasCoins, err := cosmostypes.ParseCoinsNormalized(m.faucetGas)
	if err != nil {
		return "", err
	}
	outputs := []banktypes.Output{{
		Address: toAddress,
		Coins:   gasCoins,
	}}

	inputCoins := cosmostypes.NewCoins()
	for _, o := range outputs {
		inputCoins = inputCoins.Add(o.Coins...)
	}

	gasReq := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{{
			Address: faucetAddr,
			Coins:   inputCoins,
		}},
		Outputs: outputs,
	}

	// log.Infof("Sending %s from faucet address [%s] to recipient [%s]", toAddress, faucetAddr, faucetToken)

	// Send message and get response
	res, err := m.txClient.BroadcastTx(context.Background(), *a, tokenReq, gasReq)
	if err != nil {
		return "", err
	}

	// log.Infof("tx %s from faucet address", res.TxHash)

	return res.TxHash, nil
}

// GetBalance retrieves the balance for the specified address.
func (m *Mgr) GetBalance(toAddress string) (string, error) {
	a := m.getAccount()
	if a == nil {
		return "", errors.New("no account found")
	}

	tokenBody := map[string]interface{}{
		"balance": map[string]interface{}{
			"address": toAddress,
		},
	}

	tokenJSONBody, err := json.Marshal(tokenBody)
	if err != nil {
		return "", err
	}

	tokenReq := &chaintypes.QuerySmartContractStateRequest{Address: m.tokenContract, QueryData: tokenJSONBody}

	// Send message and get response
	res, err := m.qClient.SmartContractState(context.Background(), tokenReq)
	if err != nil {
		log.Errorf("balance SmartContractState err:%s", err.Error())
		return "", err
	}

	var resp balanceResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return "", err
	}

	return resp.Balance, nil
}

type balanceResponse struct {
	Balance string `json:"balance"`
}

func (m *Mgr) sendOrder(id string, cpu, memory, disk, duration int, coin string) error {
	a := m.getAccount()
	if a == nil {
		return errors.New("no account found")
	}

	faucetAddr, err := a.Address(m.prefix)
	if err != nil {
		return err
	}

	orderBody := map[string]interface{}{
		"CreateOrder": map[string]interface{}{
			"order_id": id,
			"cpu":      cpu,
			"memory":   memory,
			"disk":     disk,
			"duration": duration * 600,
		},
	}

	orderJSONBody, err := json.Marshal(orderBody)
	if err != nil {
		return err
	}

	msg := map[string]interface{}{
		"send": map[string]interface{}{
			"amount":   coin,
			"contract": m.orderContract,
			"msg":      orderJSONBody,
		},
	}

	orderMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	orderReq := &chaintypes.MsgExecuteContract{Sender: faucetAddr, Contract: m.tokenContract, Msg: orderMsg}

	log.Infof("sendOrder %s from faucet address", faucetAddr)

	res, err := m.txClient.BroadcastTx(context.Background(), *a, orderReq)
	if err != nil {
		return err
	}
	log.Infoln(res)

	return nil
}

// TokenOrder represents an order for a token with a unique ID and duration.
type TokenOrder struct {
	ID          string `json:"id,omitempty"`
	Duration    uint64 `json:"duration,omitempty"`
	Initiator   string `json:"initiator,omitempty"`
	LockedFunds uint64 `json:"locked_funds,omitempty"`
	Resource    struct {
		CPU    uint32 `json:"cpu,omitempty"`
		Memory uint32 `json:"memory,omitempty"`
		Disk   uint32 `json:"disk,omitempty"`
	} `json:"resource"`
	StartHeight uint64      `json:"start_height,omitempty"`
	Status      OrderStatus `json:"status,omitempty"`
}

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	// Active indicates that the order is currently active.
	Active OrderStatus = "Active" // 订单活跃
	// Expired indicates that the order has expired.
	Expired OrderStatus = "Expired" // 订单到期
)

// GetOrders retrieves orders based on the provided order IDs.
func (m *Mgr) GetOrders(ids []string) ([]*TokenOrder, error) {
	a := m.getAccount()
	if a == nil {
		return nil, errors.New("no account found")
	}

	tokenBody := map[string]interface{}{
		"orders": map[string]interface{}{
			"order_ids": ids,
		},
	}

	tokenJSONBody, err := json.Marshal(tokenBody)
	if err != nil {
		return nil, err
	}

	tokenReq := &chaintypes.QuerySmartContractStateRequest{Address: m.orderContract, QueryData: tokenJSONBody}

	// Send message and get response
	res, err := m.qClient.SmartContractState(context.Background(), tokenReq)
	if err != nil {
		log.Errorf("balance SmartContractState err:%s", err.Error())
		return nil, err
	}

	var list []*TokenOrder
	err = json.Unmarshal(res.Data, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ReleaseOrder processes the release of an order by its ID.
func (m *Mgr) ReleaseOrder(id string) error {
	a := m.getAccount()
	if a == nil {
		return errors.New("no account found")
	}

	faucetAddr, err := a.Address(m.prefix)
	if err != nil {
		return err
	}

	orderBody := map[string]interface{}{
		"release_order": map[string]interface{}{
			"order_id": id,
		},
	}

	orderJSONBody, err := json.Marshal(orderBody)
	if err != nil {
		return err
	}

	orderReq := &chaintypes.MsgExecuteContract{Sender: faucetAddr, Contract: m.orderContract, Msg: orderJSONBody}

	log.Infof("ReleaseOrder %s from faucet address", faucetAddr)

	_, err = m.txClient.BroadcastTx(context.Background(), *a, orderReq)
	if err != nil {
		return err
	}
	// log.Infoln(res)

	return nil
}

// func (m *Mgr) UpdateOrder(id, newId string, cpu, memory, disk int, duration int, coin string) error {
// 	a := getAccount()
// 	if a == nil {
// 		return errors.New("no account found")
// 	}

// 	faucetAddr, err := a.Address(prefix)
// 	if err != nil {
// 		return err
// 	}

// 	orderBody := map[string]interface{}{
// 		"UpdateOrder": map[string]interface{}{
// 			"order_id":     id,
// 			"new_order_id": newId,
// 			"cpu":          cpu,
// 			"memory":       memory,
// 			"disk":         disk,
// 			"duration":     duration * 600,
// 		},
// 	}

// 	orderJSONBody, err := json.Marshal(orderBody)
// 	if err != nil {
// 		return err
// 	}

// 	msg := map[string]interface{}{
// 		"send": map[string]interface{}{
// 			"amount":   coin,
// 			"contract": orderContract,
// 			"msg":      orderJSONBody,
// 		},
// 	}

// 	orderMsg, err := json.Marshal(msg)
// 	if err != nil {
// 		return err
// 	}

// 	orderReq := &chaintypes.MsgExecuteContract{Sender: faucetAddr, Contract: tokenContract, Msg: orderMsg}

// 	log.Infof("sendOrder %s from faucet address", faucetAddr)

// 	res, err := txClient.BroadcastTx(context.Background(), *a, orderReq)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infoln(res)

// 	return nil
// }
