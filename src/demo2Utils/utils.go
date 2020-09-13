package demo2Utils

import (
	b64 "encoding/base64"
	"encoding/json"

	"go_systems/src/demo2Users"

	"golang.org/x/crypto/bcrypt"
)

func B64DecodeUser(jsonString string) ([]byte, []byte, error) {
	var u demo2Users.UserData
	if err := json.Unmarshal([]byte(jsonString), &u); err != nil {
		return nil, nil, err
	}
	emailDec, _ := b64.StdEncoding.DecodeString(string(u.Email))
	passDec, _ := b64.StdEncoding.DecodeString(string(u.Password))
	return emailDec, passDec, nil
}

func IsValid(p []byte, byteHash []byte) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(byteHash, p); err != nil {
		return false, err
	}
	return true, nil
}

func GenerateUserPassword(p string) (string, error) {
	hp, err := bcrypt.GenerateFromPassword([]byte(p), 0)
	if err != nil {
		return "", err
	}
	return string(hp), nil
}
