package yeschef

import (
	"sync"
)

type CmdServer struct {
	mu          sync.RWMutex
	Clients     map[*Client]bool
	Radio       chan []byte
	Provision   chan *Client
	Deprovision chan *Client
}

func NewServer() *CmdServer {
	return &CmdServer{
		Radio:       make(chan []byte, 1024),
		Provision:   make(chan *Client),
		Deprovision: make(chan *Client),
		Clients:     make(map[*Client]bool),
	}
}

func (xrv *CmdServer) Run() {
	for {
		select {
		case xlient := <-xrv.Provision:
			xrv.Clients[xlient] = true
		case xlient := <-xrv.Deprovision:
			if _, ok := xrv.Clients[xlient]; ok {
				delete(xrv.Clients, xlient)
				close(xlient.Xend)
			}
		case msg := <-xrv.Radio:
			for xlient := range xrv.Clients {
				select {
				case xlient.Xend <- msg:
				default:
					// Drop the message for this client if their buffer is full
				}
			}
		}
	}
}
