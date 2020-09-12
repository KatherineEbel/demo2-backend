package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go_systems/src/demo2Async"
	"go_systems/src/demo2Redis"
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
		switch in.Type {
		case "register-client":
			m := websockets.Message{
				Jwt:  "^vAr^",
				Type: "websocket-connect-success",
				Data: wsId,
			}
			if err := m.Send(c); err != nil {
				fmt.Printf("Message fail: %s", err)
			} else {
				fmt.Println("Message sent type= ", m.Type)
			}
			break
		default:
			fmt.Println("Default case")
			break
		}
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
