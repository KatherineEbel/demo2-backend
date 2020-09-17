package demo2Utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"

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

func SHA256OfString(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)
}

func GenerateUUID() (string, error) {
	f, err := os.Open("/dev/urandom")
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing ", f.Name())
		}
	}()
	if err != nil {
		return "", err
	}
	b := make([]byte, 16)
	if _, err = f.Read(b); err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}
