package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// VerifySHA256Sign verifies an RSA signature.
func VerifySHA256Sign(publicKey *rsa.PublicKey, sign []byte, content []byte) error {
	hash := crypto.SHA256.New()
	_, err := hash.Write(content)
	if err != nil {
		return err
	}

	sum := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum, sign)
	if err != nil {
		return err
	}

	return nil
}

// SHA256Sign signs a message using an RSA private key.
func SHA256Sign(privateKey *rsa.PrivateKey, content []byte) ([]byte, error) {
	hash := crypto.SHA256.New()
	_, err := hash.Write(content)
	if err != nil {
		return nil, err
	}

	sum := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Pem2PublicKey converts a PEM-encoded public key to an *rsa.PublicKey.
func Pem2PublicKey(publicPem []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicPem)
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key, %s", err.Error())
	}

	return pub, nil
}
