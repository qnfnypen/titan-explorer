package random

import (
	"math/rand"
	"time"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GenerateRandomString length is the length of random string we want to generate
func GenerateRandomString(length int) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
