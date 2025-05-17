package yeschef

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// setupTestDB initializes an isolated SQLite database for testing.
func setupTestDB(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	os.Setenv("LEMC_DATA", tmp)
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")

	mfs, err := embedded.GetMigrationsFS()
	if err != nil {
		t.Fatalf("migrations fs: %v", err)
	}
	goose.SetBaseFS(mfs)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	dbc := db.Db()
	if err := goose.Up(dbc.DB, "."); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return func() { dbc.Close() }
}

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
	teardown := setupTestDB(t)
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
	teardown := setupTestDB(t)
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
