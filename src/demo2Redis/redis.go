package demo2Redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"

	"go_systems/src/demo2Config"
)

var (
	redisClient *redis.Client
)

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: demo2Config.RedisPass,
		DB:       0,
	})
	pong, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis Connected: ", pong)
}

func RedisSet(k string, v string) error {
	return redisClient.Set(k, v, 0).Err()
}

func RedisDel(k string) error {
	_, err := redisClient.Del(k).Result()
	return err
}

type RedisTask struct {
	ws    *websocket.Conn
	kind  string
	key   string
	value string
}

func NewRedisTask(ws *websocket.Conn, tt string, tk string, tv string) *RedisTask {
	return &RedisTask{
		ws:    ws,
		kind:  tt,
		key:   tk,
		value: tv,
	}
}

func (t RedisTask) Perform() error {
	var err error
	switch t.kind {
	case "set-key":
		err = RedisSet(t.key, t.value)
		break
	case "del-key":
		fmt.Println("Setting key: ", t.key)
		err = RedisDel(t.key)
		_ = t.ws.Close()
		break
	default:
		fmt.Println("Unknown value for kind: ", t.kind)
	}
	return err
}
