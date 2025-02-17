package kub

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// GeneratePassword generates a random password of the specified length.
// It returns an error if the length is not between 8 and 64.
func (m *Mgr) GeneratePassword(length int) string {
	if length < 8 || length > 64 {
		return defaultPassword
	}

	numbers := "0123456789"
	lowerCase := "abcdefghijklmnopqrstuvwxyz"
	upperCase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	special := "~!@#$%^&*()-_=+\\|[{}];:'\",<.>/? "

	password := make([]string, 4)
	var err error

	password[0], err = randomChar(numbers)
	if err != nil {
		return defaultPassword
	}

	password[1], err = randomChar(lowerCase)
	if err != nil {
		return defaultPassword
	}

	password[2], err = randomChar(upperCase)
	if err != nil {
		return defaultPassword
	}

	password[3], err = randomChar(special)
	if err != nil {
		return defaultPassword
	}

	allChars := numbers + lowerCase + upperCase + special

	for i := 4; i < length; i++ {
		char, err := randomChar(allChars)
		if err != nil {
			return defaultPassword
		}
		password = append(password, char)
	}

	shuffled, err := shuffleSlice(password)
	if err != nil {
		return defaultPassword
	}

	return strings.Join(shuffled, "")
}

func randomChar(chars string) (string, error) {
	max := big.NewInt(int64(len(chars)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return string(chars[n.Int64()]), nil
}

func shuffleSlice(slice []string) ([]string, error) {
	result := make([]string, len(slice))
	copy(result, slice)

	for i := len(result) - 1; i > 0; i-- {
		max := big.NewInt(int64(i + 1))
		j, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, err
		}
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}

	return result, nil
}
