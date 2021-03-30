package manager

import (
	"errors"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

type (
	Client interface {
		ReadMessage(interface{}) (int, error)
		SendMessage([]byte) error
		Close()
	}

	Message struct {
		//消息struct
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
		Content   []byte `json:"content"`
	}
)

type (
	//Manager
	Manager struct {
		ClientManager
	}

	//ClientManager
	ClientManager struct {
		clients       map[string]Client
		broadcast     chan Message
		register      chan Client
		unregister    chan Client
		mux           sync.RWMutex
		SendErrHandle func(error)
	}
)

func (m *Manager) Start(done chan struct{}) {

	defer m.CloseClients()
	defer close(done)

	for {
		select {
		case message := <-m.broadcast:
			if message.Recipient == "all" {
				m.Broadcast(message)
				continue
			}

			m.Desigate(message)

		case <-done:
			break
		}
	}
}

//CloseClients
func (m *Manager) CloseClients() {

	for k, _ := range m.clients {
		m.clients[k].Close()
	}
}

func (m *Manager) Broadcast(message Message) {

	for _, v := range m.clients {
		if err := v.SendMessage(message.Content); err != nil {

			m.SendErrHandle(err)
		}
	}

}

func (m *Manager) Desigate(message Message) error {
	c, exist := m.clients[message.Recipient]
	if !exist {
		return errors.New("Non-existent connection")
	}

	return c.SendMessage(message.Content)
}

func (m *Manager) Register(id string, c Client) error {

	m.mux.Lock()

	if _, exist := m.clients[id]; exist {

		return errors.New("duplicate register")
	}
	m.clients[id] = c

	m.mux.Unlock()

	return nil
}

func (m *Manager) UnRegister(id string) {

	m.mux.Lock()
	delete(m.clients, id)
	m.mux.Unlock()
}
