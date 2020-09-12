package demo2Wsm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"go_systems/src/websockets"
)

type FatClient struct {
	Id   uuid.UUID
	Conn *websocket.Conn
	Pool *Pool
}

type SkinnyMessage struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type FatClientJson struct {
	WsId string `json:"wsId"`
}

type Pool struct {
	Register   chan *FatClient
	Unregister chan *FatClient
	Clients    map[*FatClient]bool
	Broadcast  chan SkinnyMessage
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *FatClient),
		Unregister: make(chan *FatClient),
		Clients:    make(map[*FatClient]bool),
		Broadcast:  make(chan SkinnyMessage),
	}
}

func (p *Pool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(p.Clients))
			for c := range p.Clients {
				fmt.Printf("%v\n", c)
				if err := c.Conn.WriteJSON(SkinnyMessage{
					Type: 1,
					Body: "New User Joined...",
				}); err != nil {
					fmt.Println("Register error", err)
				}
			}
			break
		case client := <-p.Unregister:
			delete(p.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(p.Clients))
			for c := range p.Clients {
				if err := c.Conn.WriteJSON(SkinnyMessage{
					Type: 1,
					Body: "User Disconnected",
				}); err != nil {
					fmt.Println("Unregister error: ", err)
				}
			}
			break
		case message := <-p.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for c := range p.Clients {
				if err := c.Conn.WriteJSON(message); err != nil {
					fmt.Println("Broadcast error: ", err)
				}
			}
		}
	}
}

func (p *Pool) NotifyWSList() {
	for range time.Tick(time.Second * 5) {
		var fcs []FatClientJson
		for client := range p.Clients {
			fcs = append(fcs, FatClientJson{WsId: client.Id.String()})
		}
		res, err := json.Marshal(fcs)
		if err == nil {
			for client := range p.Clients {
				m := &websockets.Message{
					Jwt:  strconv.Itoa(len(p.Clients)),
					Type: "fat-client-list",
					Data: string(res),
				}
				err = m.Send(client.Conn)
			}
		} else {
			fmt.Println("Marshal err ", err)
		}
	}
}
