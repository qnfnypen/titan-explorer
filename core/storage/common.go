package storage

import (
	"bytes"
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
