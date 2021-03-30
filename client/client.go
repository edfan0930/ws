package client

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
)

type Client struct {
	id            string
	socket        *websocket.Conn
	send          chan []byte
	Online        bool
	WriteDeadline int
}

//ping
func (c *Client) Ping() (err error) {

	if err = c.socket.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {

		c.Online = false

	} else if !c.Online {
		c.Online = true
	}

	return err
}

//NewClient
func NewClient(id string, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Client, error) {
	ws, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(w, r, responseHeader)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		id:     id,
		socket: ws,
		Online: true,
	}, nil
}

//WriteMessage
func (c *Client) SendMessage(message []byte) error {

	if !c.Online {
		if err := c.Ping(); err != nil {
			return err
		}
	}

	if c.WriteDeadline != 0 {
		c.socket.SetWriteDeadline(time.Now().Add(time.Duration(c.WriteDeadline) * time.Second))
	}

	err := c.socket.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		c.Close()
	}

	return err
}

//ReadMessage
func (c *Client) ReadMessage(v interface{}) (int, error) {
	messageType, message, err := c.socket.ReadMessage()

	if err != nil {

		return 0, err
	}

	if err := json.Unmarshal(message, v); err != nil {
		return 0, err
	}
	return messageType, nil
}

//Message
func (c *Client) Message(m []byte) {

	c.send <- m

	return
}

//Reader
func (c *Client) Reader() []byte {
	return <-c.send
}

//Close
func (c *Client) Close() {

	c.socket.Close()
	return
}
