package api

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/go-redis/redis/v9"
)

var AcmeRedisKey = "TITAN::CANDIDATE::ACME"

func AcmeHandler(c *gin.Context) {

	res, err := dao.RedisCache.Get(c.Request.Context(), AcmeRedisKey).Result()
	if err != nil && err != redis.Nil {
		log.Errorf("AcmeMD5Handler error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	var record *model.Acme
	if err == redis.Nil {
		record, err := dao.AcmeRecord(c.Request.Context())
		if err != nil {
			log.Errorf("AcmeMD5Handler error: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if record == nil {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}
		c.JSON(200, record)
		recordBytes, _ := json.Marshal(record)
		if err = dao.RedisCache.Set(c.Request.Context(), AcmeRedisKey, recordBytes, 0).Err(); err != nil {
			log.Errorf("AcmeMD5Handler set cache error: %v", err)
		}
		return
	}

	if err := json.Unmarshal([]byte(res), &record); err != nil {
		log.Errorf("AcmeMD5Handler error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(200, record)
}

func AcmeAddHandler(c *gin.Context) {
	crt, _, err := c.Request.FormFile("crt")
	if err != nil {
		c.String(400, "Invalid Crt")
		return
	}
	defer crt.Close()

	key, _, err := c.Request.FormFile("key")
	if err != nil {
		c.String(400, "Invalid Key")
		return
	}
	defer key.Close()

	crtBytes, err := ioutil.ReadAll(crt)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	keyBytes, err := ioutil.ReadAll(key)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	pair, err := validateKeyAndCert(keyBytes, crtBytes)
	if err != nil {
		c.String(400, "Invalid Key or Cert Pair")
		return
	}

	cert, _ := x509.ParseCertificate(pair.Certificate[0])
	expireData := cert.NotAfter

	if expireData.Unix() <= time.Now().Unix() {
		c.String(400, "Certificate expired")
		return
	}

	record := &model.Acme{
		Certificate: string(crtBytes),
		PrivateKey:  string(keyBytes),
		ExpireAt:    expireData,
		CreatedAt:   time.Now(),
	}

	if _, err := dao.RedisCache.Del(c.Request.Context(), AcmeRedisKey).Result(); err != nil {
		log.Errorf("AcmeAddHandler error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	if err := dao.AcmeAdd(c.Request.Context(), record); err != nil {
		log.Errorf("AcmeAddHandler error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, respJSON(record))
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
