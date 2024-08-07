package opcheck

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type signDocFee struct {
	Amount []sdk.Coin `json:"amount"`
	Gas    string     `json:"gas"`
}

type signDocMsgValue struct {
	Data   string `json:"data"`
	Signer string `json:"signer"`
}

type signDocMsg struct {
	Type  string          `json:"type"`
	Value signDocMsgValue `json:"value"`
}
type signDoc struct {
	AccountNumber string       `json:"account_number"`
	ChainID       string       `json:"chain_id"`
	Fee           signDocFee   `json:"fee"`
	Memo          string       `json:"memo"`
	Msgs          []signDocMsg `json:"msgs"`
	Sequence      string       `json:"sequence"`
}

// ComposeArbitraryMsg Creates SignDoc with JSON encoded bytes as per adr036
// Compatible with AMINO as it is supported by keplr wallet
func ComposeArbitraryMsg(signer string, data string) ([]byte, error) {
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(data))

	newSignDocMsgValue := signDocMsgValue{
		Data:   dataBase64,
		Signer: signer,
	}

	newSignDocMsg := signDocMsg{
		Value: newSignDocMsgValue,
		Type:  "sign/MsgSignData",
	}

	newSignDoc := signDoc{
		Msgs: []signDocMsg{
			newSignDocMsg,
		},
		AccountNumber: "0",
		Sequence:      "0",
		Fee: signDocFee{
			Gas:    "0",
			Amount: sdk.NewCoins(),
		},
	}

	jsonBytes, err := json.Marshal(newSignDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to Sign Doc to JSON: %w", err)
	}
	return jsonBytes, nil
}

// VerifyComosSign 验证 comos 链的签名
func VerifyComosSign(addr, nonce, sign, pk string) (bool, error) {
	composedArbitraryMsg, err := ComposeArbitraryMsg(addr, nonce)
	if err != nil {
		return false, fmt.Errorf("failed to compose arbitrary msg: %w", err)
	}
	bs, err := hex.DecodeString(pk)
	if err != nil {
		return false, fmt.Errorf("decode public key from pk error:%w", err)
	}
	signature, err := hex.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("decode msg of sign error:%w", err)
	}

	pubkey := &secp256k1.PubKey{Key: bs}
	return pubkey.VerifySignature(composedArbitraryMsg, signature), nil
}
