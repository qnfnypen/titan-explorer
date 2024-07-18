package storage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// CIDToHash converts a CID string to its corresponding hash string.
func CIDToHash(cidString string) (string, error) {
	cid, err := cid.Decode(cidString)
	if err != nil {
		return "", err
	}

	return cid.Hash().String(), nil
}

// HashToCID converts a hash string to its corresponding CID string.
func HashToCID(hashString string) (string, error) {
	multihash, err := mh.FromHexString(hashString)
	if err != nil {
		return "", err
	}
	cid := cid.NewCidV1(cid.Raw, multihash)
	return cid.String(), nil
}

func checkPermsIfInACL(perms []string) error {
	if len(perms) == 0 {
		return errors.New("perms can not empty")
	}

	for _, perm := range perms {
		isInACL := false
		for _, ac := range userAccessControlAll {
			if perm == ac {
				isInACL = true
				break
			}
		}

		if !isInACL {
			return fmt.Errorf("%s not in acl %s", perm, userAccessControlAll)
		}
	}

	return nil
}

// DecodeAPIKeys 解码用户 api keys
func DecodeAPIKeys(buf []byte) (map[string]UserAPIKeysInfo, error) {
	apiKeys := make(map[string]UserAPIKeysInfo)

	buffer := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&apiKeys)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// EncodeAPIKeys 编码用户 api keys
func EncodeAPIKeys(apiKeys map[string]UserAPIKeysInfo) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(apiKeys)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// DecodeAPIKeySecrets 解码用户 api key secrets
func DecodeAPIKeySecrets(buf []byte) (map[string]UserAPIKeySecretInfo, error) {
	apiKeys := make(map[string]UserAPIKeySecretInfo)

	buffer := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&apiKeys)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// EncodeAPIKeySecrets 编码用户 api key secrets
func EncodeAPIKeySecrets(apiKeys map[string]UserAPIKeySecretInfo) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(apiKeys)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ECB: 电子密码文本模式，最基本的工作模式，将待处理信息分组，每组分别进行加密或解密处理
// CBC: 密码分组链接模式，每个明文块先与其前一个密文块进行异或，然后再进行加密
// CFB: 密文反馈模式，前一个密文使用密钥key再加密后，与明文异或，得到密文。第一个密文需要初始向量IV加密得到。
// 		解密也同样使用加密器进行解密

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	// 判断缺少几位长度，最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	// 补足位数，复制padding
	padText := bytes.Repeat([]byte{(byte(padding))}, padding)

	return append(data, padText...)
}

// pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误")
	}
	// 获取填充的个数
	unPadding := int(data[length-1])

	return data[:length-unPadding], nil
}

// AesEncryptCBC 加密后再进行base64编码
func AesEncryptCBC(data []byte, key []byte) (string, error) {
	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// 判断加密块的大小
	blockSize := block.BlockSize()
	// 填充
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

// AesDecryptCBC 解密
func AesDecryptCBC(data string, key []byte) (crypted []byte, err error) {
	// 处理 panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("aes decrypt error:%v", r)
			return
		}
	}()

	// base64解码
	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 判断加密块的大小
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	crypted = make([]byte, len(dataByte))
	blockMode.CryptBlocks(crypted, dataByte)
	// 去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}

	return crypted, nil
}
