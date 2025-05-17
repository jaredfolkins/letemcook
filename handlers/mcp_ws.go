package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/yeschef"
)

var mcpUpgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

// McpWs upgrades the connection to a WebSocket implementing a basic MCP channel.
func McpWs(c LemcContext) error {
	uuid := c.Param("uuid")
	apiKey := c.Request().Header.Get("X-API-Key")
	if apiKey == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	app, perm, err := models.AppByUUIDAndUserAPIKey(uuid, apiKey)
	if err != nil {
		log.Printf("AppByUUIDAndUserAPIKey: %v", err)
		return c.NoContent(http.StatusUnauthorized)
	}
	if perm.ID == 0 {
		return c.NoContent(http.StatusUnauthorized)
	}
	if !app.IsMcpEnabled {
		return c.NoContent(http.StatusForbidden)
	}

	conn, err := mcpUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	server := yeschef.XoxoX.CreateMcpAppInstance(app.ID)
	client := &yeschef.McpClient{
		Server:    server,
		Conn:      conn,
		Send:      make(chan []byte, 64),
		UserID:    perm.UserID,
		AccountID: perm.AccountID,
	}
	server.Provision <- client
	go client.WritePump()
	go client.ReadPump()

	return nil
}
