package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go_systems/src/demo2Async"
	"go_systems/src/demo2Redis"
	"go_systems/src/websockets"
)

const (
	certPath = "/etc/letsencrypt/live/demo2.kathyebel.dev/fullchain.pem"
	keyPath  = "/etc/letsencrypt/live/demo2.kathyebel.dev/privkey.pem"
)

var (
	addr     = flag.String("addr", "0.0.0.0:1200", "http service address")
	upgrader = websocket.Upgrader{}
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
	fmt.Println(wsId)
	fmt.Println(c.LocalAddr())
	t := demo2Redis.NewRedisTask(c, "set-key", wsId, "noop")
	demo2Async.TaskQueue <- t

Loop:
	for {
		in := websockets.Message{}

		err := c.ReadJSON(&in)
		if err != nil {
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
			}
			break
		default:
			fmt.Println("Default case")
			break
		}
	}
}
func main() {
	demo2Async.StartTaskDispatcher(8)
	r := mux.NewRouter()
	r.HandleFunc("/ws", handleAPI)
	fmt.Printf("Serving TLS: %s", *addr)
	if err := http.ListenAndServeTLS(*addr, certPath, keyPath, r); err != nil {
		panic(err)
	}
}
