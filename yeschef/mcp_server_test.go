package yeschef

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// createSampleApp seeds minimal data and returns the created app and permission.
func createSampleApp(t *testing.T, yamlStr string) (*models.App, *models.PermApp) {
	dbc := db.Db()
	tx, err := dbc.Beginx()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	acc, err := models.AccountCreate("Test Account", tx)
	if err != nil {
		t.Fatalf("account create: %v", err)
	}

	pw, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	usr := models.NewUser()
	usr.Username = "tester"
	usr.Email = "tester@example.com"
	usr.Hash = string(pw)
	usr.Heckle = false
	uid, err := models.CreateUserWithAccountID(usr, acc.ID, tx)
	if err != nil {
		t.Fatalf("user create: %v", err)
	}
	usr.ID = uid

	cb := &models.Cookbook{
		AccountID:      acc.ID,
		OwnerID:        usr.ID,
		Name:           "CB",
		Description:    "desc",
		YamlShared:     yamlStr,
		YamlIndividual: yamlStr,
	}
	if err := cb.Create(tx); err != nil {
		t.Fatalf("cookbook create: %v", err)
	}

	app := &models.App{
		AccountID:      acc.ID,
		OwnerID:        usr.ID,
		CookbookID:     cb.ID,
		Name:           "App",
		Description:    "desc",
		YAMLShared:     yamlStr,
		YAMLIndividual: yamlStr,
		IsMcpEnabled:   true,
		IsActive:       true,
	}
	if err := app.Create(tx); err != nil {
		t.Fatalf("app create: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	perm, err := models.AppPermissionsByUserAccountAndApp(usr.ID, acc.ID, app.ID)
	if err != nil {
		t.Fatalf("perm fetch: %v", err)
	}

	return app, perm
}

// sampleYAML builds a small cookbook YAML with a wiki and one recipe.
func sampleYAML() string {
	yd := models.NewYamlIndividual()
	yd.Cookbook.Pages = []models.Page{{
		PageID: 1,
		Name:   "Test Page",
		Recipes: []models.Recipe{{
			Name:        "testrec",
			Description: "desc",
			Form:        []models.FormField{},
			Steps: []models.Step{{
				Step:    1,
				Image:   "docker.io/test",
				Do:      "now",
				Timeout: "1.minutes",
			}},
		}},
	}}
	yd.Cookbook.Storage.Wikis[1] = base64.StdEncoding.EncodeToString([]byte("<p>wiki</p>"))
	b, _ := yaml.Marshal(yd)
	return string(b)
}

func TestAppByUUIDAndUserAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	teardown := db.SetupTestDB(t)
	defer teardown()

	yamlStr := sampleYAML()
	app, perm := createSampleApp(t, yamlStr)

	gotApp, gotPerm, err := models.AppByUUIDAndUserAPIKey(app.UUID, perm.ApiKey)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !gotApp.IsMcpEnabled {
		t.Errorf("expected MCP enabled")
	}
	if gotPerm.ID == 0 {
		t.Errorf("expected permission record")
	}
}

func TestMcpServerPagesAndRecipes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	teardown := db.SetupTestDB(t)
	defer teardown()

	yamlStr := sampleYAML()
	app, _ := createSampleApp(t, yamlStr)

	srv := NewMcpServer(app.ID, app.UUID, yamlStr)
	client := &McpClient{Send: make(chan []byte, 2)}

	env := &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "lemc.pages"}, Client: client}
	srv.handlePages(env)
	data := <-client.Send
	var pagesResp struct {
		Result struct {
			Pages []McpPageInfo `json:"pages"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &pagesResp); err != nil {
		t.Fatalf("unmarshal pages: %v", err)
	}
	if len(pagesResp.Result.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pagesResp.Result.Pages))
	}
	pg := pagesResp.Result.Pages[0]
	if pg.Name != "Test Page" || pg.Wiki != "<p>wiki</p>" {
		t.Errorf("unexpected page data: %+v", pg)
	}
	if len(pg.Recipes) != 1 || pg.Recipes[0].Name != "testrec" {
		t.Errorf("unexpected recipes: %+v", pg.Recipes)
	}

	env = &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: "lemc.recipes"}, Client: client}
	srv.handleRecipes(env)
	data = <-client.Send
	var recResp struct {
		Result struct {
			Recipes []McpRecipeSummary `json:"recipes"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &recResp); err != nil {
		t.Fatalf("unmarshal recipes: %v", err)
	}
	if len(recResp.Result.Recipes) != 1 || recResp.Result.Recipes[0].Name != "testrec" {
		t.Errorf("unexpected recipes summary: %+v", recResp.Result.Recipes)
	}
}

func TestMcpServerApps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	teardown := db.SetupTestDB(t)
	defer teardown()

	yamlStr := sampleYAML()
	app, perm := createSampleApp(t, yamlStr)

	srv := NewMcpServer(app.ID, app.UUID, yamlStr)
	client := &McpClient{Send: make(chan []byte, 2), UserID: perm.UserID, AccountID: perm.AccountID}

	params := struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	}{Page: 1, Limit: 10}
	b, _ := json.Marshal(params)
	env := &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "lemc.apps", Params: b}, Client: client}
	srv.handleApps(env)
	data := <-client.Send
	var appsResp struct {
		Result struct {
			Apps       []McpAppInfo `json:"apps"`
			Page       int          `json:"page"`
			TotalPages int          `json:"total_pages"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &appsResp); err != nil {
		t.Fatalf("unmarshal apps: %v", err)
	}
	if len(appsResp.Result.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(appsResp.Result.Apps))
	}
	if appsResp.Result.Apps[0].UUID != app.UUID {
		t.Errorf("unexpected app uuid: %s", appsResp.Result.Apps[0].UUID)
	}
}

func TestMcpServerToolsAndResources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	teardown := db.SetupTestDB(t)
	defer teardown()

	yamlStr := sampleYAML()
	app, perm := createSampleApp(t, yamlStr)

	srv := NewMcpServer(app.ID, app.UUID, yamlStr)
	client := &McpClient{Send: make(chan []byte, 2), UserID: perm.UserID, AccountID: perm.AccountID}

	env := &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list"}, Client: client}
	srv.handleToolsList(env)
	data := <-client.Send
	var toolsResp struct {
		Result struct {
			Tools []ToolDescriptor `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &toolsResp); err != nil {
		t.Fatalf("unmarshal tools: %v", err)
	}
	if len(toolsResp.Result.Tools) != 1 || toolsResp.Result.Tools[0].Name != "run-recipe" {
		t.Fatalf("unexpected tools: %+v", toolsResp.Result.Tools)
	}

	args := struct {
		Page   int    `json:"page"`
		Recipe string `json:"recipe"`
	}{Page: 1, Recipe: "testrec"}
	ab, _ := json.Marshal(args)
	callParams := ToolCallParams{Name: "run-recipe", Arguments: ab}
	cb, _ := json.Marshal(callParams)
	env = &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: "tools/call", Params: cb}, Client: client}
	srv.handleToolsCall(env)
	<-client.Send // discard result

	env = &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`3`), Method: "resources/list"}, Client: client}
	srv.handleResourcesList(env)
	data = <-client.Send
	var resList struct {
		Result struct {
			Resources []ResourceDescriptor `json:"resources"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resList); err != nil {
		t.Fatalf("unmarshal resources: %v", err)
	}
	if len(resList.Result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resList.Result.Resources))
	}
	uri := resList.Result.Resources[0].URI

	rb, _ := json.Marshal(struct {
		URI string `json:"uri"`
	}{URI: uri})
	env = &mcpEnvelope{Msg: &McpMessage{JSONRPC: "2.0", ID: json.RawMessage(`4`), Method: "resources/read", Params: rb}, Client: client}
	srv.handleResourcesRead(env)
	data = <-client.Send
	var readResp struct {
		Result struct {
			Contents []ResourceContent `json:"contents"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &readResp); err != nil {
		t.Fatalf("unmarshal read: %v", err)
	}
	if len(readResp.Result.Contents) != 1 || readResp.Result.Contents[0].Text == "" {
		t.Fatalf("unexpected read result: %+v", readResp.Result.Contents)
	}
}
