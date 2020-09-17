package demo2Config

import (
	"crypto/rsa"
	"fmt"
	"os"

	jwtGo "github.com/dgrijalva/jwt-go"

	"go_systems/src/demo2fs"
)

const (
	FileStoragePath = "/var/www/uploads/"
)

var (
	MySqlUser = os.Getenv("MYSQL_USER")
	MySqlPass = os.Getenv("MYSQL_PASS")
	MongoHost = os.Getenv("MONGO_HOST")
	MongoUser = os.Getenv("MONGO_USER")
	MongoPass = os.Getenv("MONGO_PASS")
	MongoDB   = os.Getenv("MONGO_DB")
	RedisPass = os.Getenv("REDIS_PASS")
	PKPass    = os.Getenv("PK_PASS")

	KeyCertPath = os.Getenv("KEY_CERT_PATH")
	PrivKeyPath = os.Getenv("PRIV_KEY_PATH")
	PubKeyPath  = os.Getenv("PUB_KEY_PATH")

	PubKeyFile  *rsa.PublicKey
	PrivKeyFile *rsa.PrivateKey
)

func loadPublicKey() {
	f, err := demo2fs.ReadFile(PubKeyPath)
	if err != nil {
		panic(err)
	}
	PubKeyFile, err = jwtGo.ParseRSAPublicKeyFromPEM(f)
	if err != nil {
		panic(err)
	}
	fmt.Println("Pub key loaded")
}

func loadPrivateKey() {
	f, err := demo2fs.ReadFile(PrivKeyPath)
	if err != nil {
		panic(err)
	}
	PrivKeyFile, err = jwtGo.ParseRSAPrivateKeyFromPEMWithPassword(f, PKPass)
	if err != nil {
		panic(err)
	}
	fmt.Println("Priv key loaded")
}
func init() {
	loadPublicKey()
	loadPrivateKey()
	fmt.Println("Pub and Priv keys loaded...")
}
