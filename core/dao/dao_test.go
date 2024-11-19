package dao

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/Masterminds/squirrel"
)

func TestMain(m *testing.M) {
	// db, err := sqlx.Connect("mysql", "root:abcd1234@tcp(localhost:8080)/titan_explorer?charset=utf8mb4&parseTime=true&loc=Local")
	// if err != nil {
	// 	panic(err)
	// }

	// DB = db

	m.Run()
}

func TestMoveBackDeletedDevice(t *testing.T) {
	err := MoveBackDeletedDevice(context.Background(), []string{"1"}, "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetNodeNums(t *testing.T) {
	on, ab, off, del, err := GetNodeNums(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(on, ab, off, del)
}

func TestDeleteUserGroupAsset(t *testing.T) {
	ctx := context.Background()
	userID := "0x5e48ee53a85343b7b57014a1eb20e21fff92d4a4"
	gids := []int64{}

	err := DeleteUserGroupAsset(ctx, userID, gids)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	userID := "0x5e48ee53a85343b7b57014a1eb20e21fff92d4a4"
	query, args, err := squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("GREATEST(used_storage_size - ?,0)", 1000)).Where("username = ?", userID).ToSql()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(query, args)
}

func TestAce(t *testing.T) {
	// crt, _ := os.ReadFile("/Users/zt/Documents/_test.titannet.io.crt")
	// key, _ := os.ReadFile("/Users/zt/Documents/_test.titannet.io.key")

	// pair, err := validateKeyAndCert(key, crt)
	// if err != nil {
	// 	panic(err)
	// }

	// cert, _ := x509.ParseCertificate(pair.Certificate[0])
	// expireData := cert.NotAfter
	// v := &model.Acme{
	// 	Certificate: string(crt),
	// 	PrivateKey:  string(key),
	// 	CreatedAt:   time.Now(),
	// 	ExpireAt:    expireData,
	// }
	// vs, _ := json.Marshal(v)
	// fmt.Println(string(vs))
	c, err := rand.Int(rand.Reader, big.NewInt(1e16))
	if err != nil {
		panic(err)
	}
	nonce := fmt.Sprintf("%d", c)

	t.Log(nonce)
}

var InvalidCertKeyPaid = errors.New("Invalid private key file or cert file")

// validateKeyAndCert validates the provided key and certificate.
func validateKeyAndCert(keyBytes, crtBytes []byte) (*tls.Certificate, error) {
	// Decode the private key.

	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, InvalidCertKeyPaid
	}
	var privateKey crypto.PrivateKey
	var err error
	switch keyBlock.Type {
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		return nil, errors.New("Unsupported private key type")
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}

	// Decode the certificate.
	certBlock, _ := pem.Decode(crtBytes)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, InvalidCertKeyPaid
	}
	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse certificate: %v", err)
	}

	// Verify the private key matches the certificate.
	certPrivateKey, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, InvalidCertKeyPaid
	}

	// Create a test TLS certificate.
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certificate.Raw},
		PrivateKey:  certPrivateKey,
	}

	// Check if the certificate can be used to create a TLS connection.
	_, err = tls.X509KeyPair(pem.EncodeToMemory(certBlock), pem.EncodeToMemory(keyBlock))
	if err != nil {
		return nil, fmt.Errorf("Certificate and key do not match: %v", err)
	}

	return &tlsCert, nil
}
