package web

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	SIGNING_KEY = "signing-key"
)

func CreateClaims() (string, error) {
	key := []byte(SIGNING_KEY)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(1962978897, 0)),
		Issuer:    "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(key)

	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return ss, nil
}
