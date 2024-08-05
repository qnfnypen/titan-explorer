package rsa

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var FilePassSalt = "TITAN_NETWORK_FILE_PASS"

// VerifyAddrSign 判断以太坊地址的签名是否正确
func VerifyAddrSign(nonce, sign string) (string, error) {
	// Hash the unsigned message using EIP-191
	hashedMessage := []byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(nonce)) + nonce)
	hash := crypto.Keccak256Hash(hashedMessage)
	// Get the bytes of the signed message
	decodedMessage := hexutil.MustDecode(sign)
	// Handles cases where EIP-115 is not implemented (most wallets don't implement it)
	if decodedMessage[64] == 27 || decodedMessage[64] == 28 {
		decodedMessage[64] -= 27
	}
	// Recover a public key from the signed message
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), decodedMessage)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(*sigPublicKeyECDSA).String(), nil
}

func EncryptPassWithSalt(pass string) string {
	bs := crypto.Keccak256([]byte(pass + FilePassSalt))
	return hexutil.Encode(bs)
}
