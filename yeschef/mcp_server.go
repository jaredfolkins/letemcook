package yeschef

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// McpMessage represents a generic MCP JSON-RPC message.
type McpMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// McpClient is a WebSocket client connected to an McpServer.
type McpClient struct {
	Server *McpServer
	Conn   *websocket.Conn
	Send   chan []byte
}

func (c *McpClient) ReadPump() {
	defer func() {
		c.Server.Deprovision <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(MAX_MESSAGE_SIZE)
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg McpMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			c.Server.Inbound <- &msg
		}
	}
}

func (c *McpClient) WritePump() {
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// McpServer manages a set of McpClients.
type McpServer struct {
	mu          sync.RWMutex
	Clients     map[*McpClient]bool
	Inbound     chan *McpMessage
	Provision   chan *McpClient
	Deprovision chan *McpClient
}

func NewMcpServer() *McpServer {
	return &McpServer{
		Clients:     make(map[*McpClient]bool),
		Inbound:     make(chan *McpMessage, 64),
		Provision:   make(chan *McpClient),
		Deprovision: make(chan *McpClient),
	}
}

func (srv *McpServer) Run() {
	for {
		select {
		case c := <-srv.Provision:
			srv.Clients[c] = true
		case c := <-srv.Deprovision:
			if _, ok := srv.Clients[c]; ok {
				delete(srv.Clients, c)
				close(c.Send)
			}
		case msg := <-srv.Inbound:
			log.Printf("MCP message received: method=%s", msg.Method)
			// Placeholder: real MCP handling would go here
		}
	}
}
