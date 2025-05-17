package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/yeschef"
)

// McpSSE establishes an SSE stream for MCP messages.
func McpSSE(c LemcContext) error {
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

	w := c.Response()
	flusher, ok := w.Writer.(http.Flusher)
	if !ok {
		return c.NoContent(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	server := yeschef.XoxoX.CreateMcpAppInstance(app.ID)
	client := &yeschef.McpClient{
		Server:    server,
		Send:      make(chan []byte, 64),
		UserID:    perm.UserID,
		AccountID: perm.AccountID,
		ApiKey:    apiKey,
	}
	server.Provision <- client

	ctx := c.Request().Context()
	go func() {
		<-ctx.Done()
		server.Deprovision <- client
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			if !ok {
				return nil
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return nil
		}
	}
}

// McpPost handles MCP JSON-RPC requests sent via POST.
func McpPost(c LemcContext) error {
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

	server := yeschef.XoxoX.CreateMcpAppInstance(app.ID)
	client := server.FindClient(apiKey)
	if client == nil {
		return c.NoContent(http.StatusGone)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var msg yeschef.McpMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	server.Enqueue(client, &msg)
	return c.NoContent(http.StatusAccepted)
}
