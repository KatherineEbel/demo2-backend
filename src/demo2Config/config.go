package demo2Config

import (
	"os"
)

var (
	RedisPass = os.Getenv("REDIS_PASS")
)

func init() {

}
