package websockets

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Jwt  string `json:"jwt"`
	Type string `json:"type"`
	Data string `json:"data"`
	mu   sync.Mutex
}

type RestMessage struct {
	Jwt  string `json:"jwt"`
	Type string `json:"type"`
	Data string `json:"data"`
}

func (m *Message) Send(c *websocket.Conn) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return c.WriteJSON(m)
}
