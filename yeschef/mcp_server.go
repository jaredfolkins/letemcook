package yeschef

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"gopkg.in/yaml.v3"
)

const defaultAppsLimit = 10

// McpMessage represents a generic MCP JSON-RPC message.
type McpMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// McpClient is a WebSocket client connected to an McpServer.
type McpClient struct {
	Server    *McpServer
	Conn      *websocket.Conn
	Send      chan []byte
	UserID    int64
	AccountID int64
	ApiKey    string
}

type mcpEnvelope struct {
	Msg    *McpMessage
	Client *McpClient
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
			c.Server.Inbound <- &mcpEnvelope{Msg: &msg, Client: c}
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
	Inbound     chan *mcpEnvelope
	Provision   chan *McpClient
	Deprovision chan *McpClient
	AppUUID     string
	AppID       int64
	YAML        string
	Tools       []ToolDescriptor
}

func NewMcpServer(appID int64, uuid string, yamlStr string) *McpServer {
	srv := &McpServer{
		Clients:     make(map[*McpClient]bool),
		Inbound:     make(chan *mcpEnvelope, 64),
		Provision:   make(chan *McpClient),
		Deprovision: make(chan *McpClient),
		AppUUID:     uuid,
		AppID:       appID,
		YAML:        yamlStr,
	}
	srv.Tools = []ToolDescriptor{
		{
			Name:        "run-recipe",
			Description: "Run a recipe by page and name",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":   map[string]interface{}{"type": "integer"},
					"recipe": map[string]interface{}{"type": "string"},
				},
				"required": []string{"page", "recipe"},
			},
		},
	}
	return srv
}

func (srv *McpServer) broadcast(b []byte) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	for c := range srv.Clients {
		select {
		case c.Send <- b:
		default:
		}
	}
}

// FindClient returns the first client associated with the given API key.
func (srv *McpServer) FindClient(apiKey string) *McpClient {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	for c := range srv.Clients {
		if c.ApiKey == apiKey {
			return c
		}
	}
	return nil
}

// Enqueue sends a message from a client to the server.
func (srv *McpServer) Enqueue(c *McpClient, msg *McpMessage) {
	srv.Inbound <- &mcpEnvelope{Msg: msg, Client: c}
}

type McpRecipeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

type McpRecipeSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type McpPageInfo struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Wiki    string          `json:"wiki"`
	Recipes []McpRecipeInfo `json:"recipes"`
}

type McpAppInfo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

// ToolDescriptor describes a single MCP tool.
type ToolDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ToolCallParams represents parameters for tools/call.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ResourceDescriptor describes a resource for resources/list.
type ResourceDescriptor struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent represents the content returned by resources/read.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

func (srv *McpServer) handleMessage(env *mcpEnvelope) {
	switch env.Msg.Method {
	case "lemc.pages":
		srv.handlePages(env)
	case "lemc.recipes":
		srv.handleRecipes(env)
	case "lemc.apps":
		srv.handleApps(env)
	case "lemc.run":
		srv.handleRun(env)
	case "tools/list":
		srv.handleToolsList(env)
	case "tools/call":
		srv.handleToolsCall(env)
	case "resources/list":
		srv.handleResourcesList(env)
	case "resources/read":
		srv.handleResourcesRead(env)
	default:
		log.Printf("MCP unknown method: %s", env.Msg.Method)
	}
}

func (srv *McpServer) handlePages(env *mcpEnvelope) {
	var yd models.YamlDefault
	if err := yaml.Unmarshal([]byte(srv.YAML), &yd); err != nil {
		srv.sendError(env, fmt.Sprintf("yaml: %v", err))
		return
	}
	yd.UUID = srv.AppUUID
	var pages []McpPageInfo
	for _, p := range yd.Cookbook.Pages {
		pi := McpPageInfo{ID: p.PageID, Name: p.Name}
		if w, ok := yd.Cookbook.Storage.Wikis[p.PageID]; ok {
			if dec, err := base64.StdEncoding.DecodeString(w); err == nil {
				pi.Wiki = string(dec)
			}
		}
		for _, r := range p.Recipes {
			action := fmt.Sprintf("/lemc/app/job/shared/uuid/%s/page/%d/recipe/%s", srv.AppUUID, p.PageID, r.Name)
			pi.Recipes = append(pi.Recipes, McpRecipeInfo{Name: r.Name, Description: r.Description, Action: action})
		}
		pages = append(pages, pi)
	}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"pages": pages}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleRecipes(env *mcpEnvelope) {
	var yd models.YamlDefault
	if err := yaml.Unmarshal([]byte(srv.YAML), &yd); err != nil {
		srv.sendError(env, fmt.Sprintf("yaml: %v", err))
		return
	}
	var recipes []McpRecipeSummary
	for _, p := range yd.Cookbook.Pages {
		for _, r := range p.Recipes {
			recipes = append(recipes, McpRecipeSummary{Name: r.Name, Description: r.Description})
		}
	}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"recipes": recipes}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleApps(env *mcpEnvelope) {
	var params struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	}
	if err := json.Unmarshal(env.Msg.Params, &params); err != nil {
		srv.sendError(env, fmt.Sprintf("params: %v", err))
		return
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = defaultAppsLimit
	}

	userID := env.Client.UserID
	accountID := env.Client.AccountID

	total, err := models.Countapps(userID, accountID)
	if err != nil {
		srv.sendError(env, err.Error())
		return
	}
	apps, err := models.Apps(userID, accountID, params.Page, params.Limit)
	if err != nil {
		srv.sendError(env, err.Error())
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.Limit)))
	}
	var infos []McpAppInfo
	for _, a := range apps {
		infos = append(infos, McpAppInfo{UUID: a.UUID, Name: a.Name, Description: a.Description})
	}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"apps": infos, "page": params.Page, "total_pages": totalPages}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleRun(env *mcpEnvelope) {
	var params struct {
		Page   int    `json:"page"`
		Recipe string `json:"recipe"`
	}
	if err := json.Unmarshal(env.Msg.Params, &params); err != nil {
		srv.sendError(env, fmt.Sprintf("params: %v", err))
		return
	}
	if err := srv.runRecipe(params.Page, params.Recipe); err != nil {
		srv.sendError(env, err.Error())
		return
	}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: "ok"}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleToolsList(env *mcpEnvelope) {
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"tools": srv.Tools}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleToolsCall(env *mcpEnvelope) {
	var params ToolCallParams
	if err := json.Unmarshal(env.Msg.Params, &params); err != nil {
		srv.sendError(env, fmt.Sprintf("params: %v", err))
		return
	}
	switch params.Name {
	case "run-recipe":
		var args struct {
			Page   int    `json:"page"`
			Recipe string `json:"recipe"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			srv.sendError(env, fmt.Sprintf("args: %v", err))
			return
		}
		if err := srv.runRecipe(args.Page, args.Recipe); err != nil {
			srv.sendError(env, err.Error())
			return
		}
		result := map[string]interface{}{
			"content": []map[string]string{{"type": "text", "text": "ok"}},
		}
		resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: result}
		b, _ := json.Marshal(resp)
		env.Client.Send <- b
	default:
		srv.sendError(env, "unknown tool")
	}
}

func (srv *McpServer) handleResourcesList(env *mcpEnvelope) {
	var yd models.YamlDefault
	if err := yaml.Unmarshal([]byte(srv.YAML), &yd); err != nil {
		srv.sendError(env, fmt.Sprintf("yaml: %v", err))
		return
	}
	var resources []ResourceDescriptor
	for _, p := range yd.Cookbook.Pages {
		if w, ok := yd.Cookbook.Storage.Wikis[p.PageID]; ok {
			uri := fmt.Sprintf("lemc://app/%s/wiki/%d", srv.AppUUID, p.PageID)
			resources = append(resources, ResourceDescriptor{URI: uri, Name: fmt.Sprintf("Page %d Wiki", p.PageID), MimeType: "text/html"})
		}
	}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"resources": resources}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) handleResourcesRead(env *mcpEnvelope) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(env.Msg.Params, &params); err != nil {
		srv.sendError(env, fmt.Sprintf("params: %v", err))
		return
	}
	wikiRegex := regexp.MustCompile(`^lemc://app/[^/]+/wiki/(\d+)$`)
	matches := wikiRegex.FindStringSubmatch(params.URI)
	if len(matches) != 2 {
		srv.sendError(env, "unknown resource")
		return
	}
	pageID, _ := strconv.Atoi(matches[1])
	var yd models.YamlDefault
	if err := yaml.Unmarshal([]byte(srv.YAML), &yd); err != nil {
		srv.sendError(env, fmt.Sprintf("yaml: %v", err))
		return
	}
	w, ok := yd.Cookbook.Storage.Wikis[pageID]
	if !ok {
		srv.sendError(env, "resource not found")
		return
	}
	dec, err := base64.StdEncoding.DecodeString(w)
	if err != nil {
		srv.sendError(env, fmt.Sprintf("decode: %v", err))
		return
	}
	content := ResourceContent{URI: params.URI, MimeType: "text/html", Text: string(dec)}
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Result: map[string]interface{}{"contents": []ResourceContent{content}}}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
}

func (srv *McpServer) runRecipe(page int, recipeName string) error {
	var yd models.YamlDefault
	if err := yaml.Unmarshal([]byte(srv.YAML), &yd); err != nil {
		return fmt.Errorf("yaml: %v", err)
	}
	var rec models.Recipe
	found := false
	for _, p := range yd.Cookbook.Pages {
		if p.PageID == page {
			for _, r := range p.Recipes {
				if r.Name == recipeName {
					rec = r
					found = true
					break
				}
			}
		}
	}
	if !found {
		return fmt.Errorf("recipe not found")
	}

	var envVars []string
	envVars = append(envVars, yd.Cookbook.Environment.Private...)
	envVars = append(envVars, yd.Cookbook.Environment.Public...)
	envVars = append(envVars, fmt.Sprintf("LEMC_STEP_ID=%d", 1))
	envVars = append(envVars, "LEMC_SCOPE=shared")
	envVars = append(envVars, fmt.Sprintf("LEMC_UUID=%s", srv.AppUUID))
	envVars = append(envVars, fmt.Sprintf("LEMC_RECIPE_NAME=%s", util.AlphaNumHyphen(rec.Name)))
	envVars = append(envVars, fmt.Sprintf("LEMC_PAGE_ID=%d", page))

	jr := &JobRecipe{
		JobType:  JOB_TYPE_APP,
		UUID:     srv.AppUUID,
		AppID:    fmt.Sprintf("%d", srv.AppID),
		PageID:   fmt.Sprintf("%d", page),
		UserID:   "0",
		Username: "mcp",
		Scope:    "shared",
		Env:      envVars,
		Recipe:   rec,
	}

	srv.broadcast([]byte("--MCP JOB STARTED--"))
	if err := DoNow(jr); err != nil {
		return err
	}
	srv.broadcast([]byte("--MCP JOB FINISHED--"))
	return nil
}

func (srv *McpServer) sendError(env *mcpEnvelope, msg string) {
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: env.Msg.ID, Error: msg}
	b, _ := json.Marshal(resp)
	env.Client.Send <- b
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
		case env := <-srv.Inbound:
			srv.handleMessage(env)
		}
	}
}
