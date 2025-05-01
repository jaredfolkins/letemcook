package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jaredfolkins/letemcook/yeschef"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Ws(c LemcContext) error {
	userCtx := c.UserContext()

	if userCtx == nil || userCtx.ActingAs == nil || userCtx.ActingAs.Account == nil {
		log.Println("User is not authenticated")
		// Use a more specific status code if needed
		return c.NoContent(http.StatusUnauthorized)
	}

	userID := userCtx.ActingAs.ID

	xonn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket for user %d: %v", userID, err)
		return err
	}

	log.Printf("Creating websocket client for user %d", userID)
	xerver := yeschef.XoxoX.CreadInstance(userID)

	// It's good practice to check if CreadInstance somehow failed, though it currently doesn't return nil
	if xerver == nil {
		log.Printf("Failed to get CmdServer instance for user %d", userID)
		xonn.Close()
		return c.NoContent(http.StatusInternalServerError)
	}

	xlient := &yeschef.Client{
		Xend:   make(chan []byte, 1024),
		Xerver: xerver,
		Xonn:   xonn,
	}

	// Register the client with the server
	xlient.Xerver.Provision <- xlient

	// Start the read/write pumps
	go xlient.WritePump()
	go xlient.ReadPump()

	return nil // Indicate successful handling
}
