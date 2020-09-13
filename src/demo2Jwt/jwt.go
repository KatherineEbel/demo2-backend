package demo2Jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwtGo "github.com/dgrijalva/jwt-go"
)

func GenerateJwt(key *rsa.PrivateKey) (string, error) {
	token := jwtGo.New(jwtGo.SigningMethodRS256)
	in60min := time.Now().Add(time.Hour).Unix()
	token.Claims = jwtGo.MapClaims{
		"iss":    "demo2.kathyebel.dev",
		"aud":    "demo2.kathyebel.dev",
		"exp":    in60min,
		"jti":    "Unique",
		"iat":    time.Now().Unix(),
		"nbf":    2,
		"sub":    "subject",
		"scopes": "api:read",
	}
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJwt(key *rsa.PublicKey, jwt string) (bool, error) {
	token, err := jwtGo.Parse(jwt, func(token *jwtGo.Token) (interface{}, error) {
		return key, nil
	})
	if token != nil && token.Valid {
		return token.Valid, nil
	} else if ve, ok := err.(*jwtGo.ValidationError); ok {
		if ve.Errors&jwtGo.ValidationErrorMalformed != 0 {
			return false, errors.New("invalid value for token")
		} else if ve.Errors&(jwtGo.ValidationErrorExpired|jwtGo.ValidationErrorNotValidYet) != 0 {
			return false, errors.New("token is expired")
		} else {
			return false, errors.New(fmt.Sprintf("couldn't handle this token: %s", err))
		}
	}
	return false, err
}
