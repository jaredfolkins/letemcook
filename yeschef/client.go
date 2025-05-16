package yeschef

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Xerver *CmdServer
	Xonn   *websocket.Conn
	Xend   chan []byte
	mu     sync.RWMutex
}

func (x *Client) ReadPump() {
	defer func() {
		x.Xerver.Deprovision <- x
		x.Xonn.Close()
	}()
	x.Xonn.SetReadLimit(MAX_MESSAGE_SIZE)
	x.Xonn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	x.Xonn.SetPongHandler(func(string) error { x.Xonn.SetReadDeadline(time.Now().Add(PONG_WAIT)); return nil })
	for {
		_, message, err := x.Xonn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		x.Xerver.Radio <- message
	}
}

func (x *Client) WritePump() {
	x.Xonn.EnableWriteCompression(true)
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		x.Xonn.Close()
	}()
	for {
		select {
		case message, ok := <-x.Xend:
			x.Xonn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if !ok {
				x.Xonn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := x.Xonn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(x.Xend)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-x.Xend)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			x.Xonn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := x.Xonn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
