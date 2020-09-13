package demo2KeyGen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
)

func GenerateKeyPair(bits int, pwd string) (privKeyBytes []byte, pubKeyBytes []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	if err = key.Validate(); err != nil {
		log.Fatal(err.Error())
	}
	pubKeyBytes, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privKeyBlk := &pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(key),
	}

	pubKeyBlock := &pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   pubKeyBytes,
	}
	if pwd != "" {
		privKeyBlk, err = x509.EncryptPEMBlock(rand.Reader, privKeyBlk.Type, privKeyBlk.Bytes, []byte(pwd), x509.PEMCipherAES256)
		if err != nil {
			return nil, nil, err
		}
	}
	privKeyBytes = pem.EncodeToMemory(privKeyBlk)
	pubKeyBytes = pem.EncodeToMemory(pubKeyBlock)
	return privKeyBytes, pubKeyBytes, nil
}
