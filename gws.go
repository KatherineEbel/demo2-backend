package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go_systems/src/demo2Async"
	"go_systems/src/demo2Config"
	"go_systems/src/demo2Jwt"
	"go_systems/src/demo2Mongo"
	"go_systems/src/demo2Redis"
	"go_systems/src/demo2Utils"
	"go_systems/src/demo2Wsm"
	"go_systems/src/websockets"
)

const (
	certPath = "/etc/letsencrypt/live/demo2.kathyebel.dev/fullchain.pem"
	keyPath  = "/etc/letsencrypt/live/demo2.kathyebel.dev/privkey.pem"
)

var (
	addr     = flag.String("addr", "0.0.0.0:1200", "http service address")
	upgrader = websocket.Upgrader{}
	pool     = demo2Wsm.NewPool()
)

func sendJWTErr(c *websocket.Conn, err error) error {
	m := &websockets.Message{
		Jwt:  "noop",
		Type: "invalid-credentials",
		Data: err.Error(),
	}
	return m.Send(c)
}

func handleMessage(message *websockets.Message, c *websocket.Conn) {
	switch message.Type {
	case "get-jwt":
		m, err := handleGetJWT(message.Data)
		if err != nil {
			fmt.Println(sendJWTErr(c, err))
			break
		}
		fmt.Println(m.Send(c))
		break
	case "validate-jwt":
		ok, err := demo2Jwt.ValidateJwt(demo2Config.PubKeyFile, message.Jwt)
		if err != nil {
			err = errors.New("server error while validating request")
			fmt.Println(sendJWTErr(c, err))
		}
		d := struct {
			Valid bool
		}{ok}
		md, err := json.Marshal(d)
		m := &websockets.Message{
			Jwt:  message.Jwt,
			Type: "jwt-valid",
			Data: string(md),
		}
		fmt.Println(m.Send(c))
		break
	default:
		fmt.Println("Got unknown message type: ", message.Type)
	}
}
func handleGetJWT(data string) (*websockets.Message, error) {
	email, password, err := demo2Utils.B64DecodeUser(data)
	if err != nil {
		err = errors.New("invalid user data")
		return nil, err
	}
	user, err := demo2Mongo.AuthenticateUser(email, password)
	if err != nil {
		err = errors.New("invalid user credentials")
		return nil, err
	}
	user.Password = "foo"
	d, err := json.Marshal(user)
	if err != nil {
		err = errors.New("server error")
		return nil, err
	}
	jwt, err := demo2Jwt.GenerateJwt(demo2Config.PrivKeyFile)
	if err != nil {
		err = errors.New("server error creating JWT")
		return nil, err
	}
	m := &websockets.Message{
		Jwt:  jwt,
		Type: "jwt-token",
		Data: string(d),
	}
	return m, nil
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WTF handleAPI WS Upgrade Err: %s", err)
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Error creating id: %s", err)
	}
	wsId := "ws-" + id.String()
	fmt.Println("wsID: ", wsId)
	fmt.Println("conn local-addr ", c.LocalAddr())
	t := demo2Redis.NewRedisTask(c, "set-key", wsId, "noop")
	demo2Async.TaskQueue <- t
	fc := &demo2Wsm.FatClient{
		Id:   id,
		Conn: c,
		Pool: pool,
	}
	pool.Register <- fc
	m := &websockets.Message{
		Jwt:  "^vAr^",
		Type: "client-websocket-id",
		Data: id.String(),
	}
	if err := m.Send(c); err != nil {
		fmt.Println("error sending message", err)
	}
Loop:
	for {
		in := websockets.Message{}

		err := c.ReadJSON(&in)
		if err != nil {
			pool.Unregister <- fc
			rt := demo2Redis.NewRedisTask(c, "del-key", wsId, "noop")
			demo2Async.TaskQueue <- rt
			_ = c.Close()
			break Loop
		}
		handleMessage(&in, c)

	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	go demo2Async.StartTaskDispatcher(9)
	go pool.Start()
	go pool.NotifyWSList()
	r := mux.NewRouter()
	r.HandleFunc("/ws", handleAPI)
	fmt.Printf("Serving TLS: %s\n", *addr)
	if err := http.ListenAndServeTLS(*addr, certPath, keyPath, r); err != nil {
		panic(err)
	}
}
