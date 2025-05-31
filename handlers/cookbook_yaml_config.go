package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/jaredfolkins/letemcook/views/partials"
	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
)

func GetCookbookYamlConfigHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	uuid := c.Param("uuid")
	view_type := c.Param("view_type")
	err := cb.ByUUID(uuid)
	if err != nil {
		return err
	}

	var yaml_default models.YamlDefault
	var yaml_default_no_storage models.YamlDefaultNoStorage

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		err = yaml.Unmarshal([]byte(cb.YamlIndividual), &yaml_default)
		if err != nil {
			return err
		}
	case SCOPE_YAML_TYPE_SHARED:
		err = yaml.Unmarshal([]byte(cb.YamlShared), &yaml_default)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("view_type not found")
	}

	// Reorder pages sequentially if needed
	yaml_default.Cookbook.Pages = models.ReorderPagesSequentially(yaml_default.Cookbook.Pages)

	yaml_default_no_storage.Cookbook.Pages = yaml_default.Cookbook.Pages
	yaml_default_no_storage.Cookbook.Environment = yaml_default.Cookbook.Environment

	v := models.CoreView{
		Cookbook:             cb,
		YamlDefaultNoStorage: yaml_default_no_storage,
		ViewType:             view_type,
		BaseView:             NewBaseView(c),
	}

	re := pages.RenderYaml(v)
	return HTML(c, re)
}

// Helper function to load YAML data based on view type
func loadYamlByViewType(cb *models.Cookbook, viewType string) (*models.YamlDefault, error) {
	var yamlDefault models.YamlDefault
	var err error

	switch viewType {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		err = yaml.Unmarshal([]byte(cb.YamlIndividual), &yamlDefault)
	case SCOPE_YAML_TYPE_SHARED:
		err = yaml.Unmarshal([]byte(cb.YamlShared), &yamlDefault)
	default:
		return nil, fmt.Errorf("view_type not found")
	}

	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}

	return &yamlDefault, nil
}

// Helper function to process form data into YAML
func processFormData(formData string, action string) (*models.YamlDefaultNoStorage, error) {
	var yamlNoStorage models.YamlDefaultNoStorage

	switch action {
	case "all":
		err := yaml.Unmarshal([]byte(formData), &yamlNoStorage)
		if err != nil {
			return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}

	return &yamlNoStorage, nil
}

// Helper function to save YAML data back to cookbook
func saveYamlToCookbook(cb *models.Cookbook, yamlDefault *models.YamlDefault, viewType string) (bool, error) {
	prettyYAML, err := yaml.Marshal(yamlDefault)
	if err != nil {
		return false, fmt.Errorf("yaml marshal failed: %w", err)
	}

	var isAdmin bool
	switch viewType {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		cb.YamlIndividual, err = prettyPrintYAML(string(prettyYAML))
		if err != nil {
			return false, fmt.Errorf("yaml pretty print failed: %w", err)
		}
	case SCOPE_YAML_TYPE_SHARED:
		cb.YamlShared, err = prettyPrintYAML(string(prettyYAML))
		if err != nil {
			return false, fmt.Errorf("yaml pretty print failed: %w", err)
		}
		isAdmin = true
	default:
		return false, fmt.Errorf("view_type not found")
	}

	return isAdmin, nil
}

// Helper function to process pages and cache generation
func processPages(yamlDefault *models.YamlDefault, cb *models.Cookbook, viewType string, userContext *models.UserContext, isAdmin bool) error {
	yamlDefault.UUID = cb.UUID

	for _, page := range yamlDefault.Cookbook.Pages {
		if err := processPage(page, cb, viewType, userContext, isAdmin); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to process individual page
func processPage(page models.Page, cb *models.Cookbook, viewType string, userContext *models.UserContext, isAdmin bool) error {
	jm := &util.JobMeta{
		UUID:     cb.UUID,
		PageID:   strconv.Itoa(page.PageID),
		UserID:   fmt.Sprintf("%d", userContext.ActingAs.Account.ID),
		Username: userContext.ActingAs.Username,
	}

	cf, err := util.NewContainerFiles(jm, isAdmin)
	if err != nil {
		return fmt.Errorf("container files creation failed: %w", err)
	}

	if err := cf.OpenFiles(); err != nil {
		return fmt.Errorf("opening files failed: %w", err)
	}
	defer cf.CloseFiles()

	// Generate CSS cache
	css, err := cf.Read(cf.Css)
	if err != nil {
		return fmt.Errorf("reading CSS failed: %w", err)
	}
	page.CssCache = fmt.Sprintf("<style id='uuid-%s-pageid-%d-scope-%s-style'>%s</style>", cb.UUID, page.PageID, viewType, css)

	// Generate HTML cache
	html, err := cf.Read(cf.Html)
	if err != nil {
		return fmt.Errorf("reading HTML failed: %w", err)
	}
	page.HtmlCache = fmt.Sprintf("<div id='uuid-%s-pageid-%d-scope-%s-html' class='page-inner'>%s</div>", cb.UUID, page.PageID, viewType, html)

	// Generate JS cache
	js, err := cf.Read(cf.Js)
	if err != nil {
		return fmt.Errorf("reading JS failed: %w", err)
	}
	page.JsCache = fmt.Sprintf("<script id='uuid-%s-pageid-%d-scope-%s-script'>%s</script>", cb.UUID, page.PageID, viewType, js)

	return nil
}

func PostCookbookYamlConfigHandler(c LemcContext) error {
	// Extract parameters
	unsafeUuid := c.Param("uuid")
	unsafeAction := c.Param("action")
	viewType := c.Param("view_type")
	formData := c.FormValue(fmt.Sprintf("yaml-%s", unsafeAction))

	// Load cookbook
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	if err := cb.ByUUID(unsafeUuid); err != nil {
		log.Printf("ByUUID failed: %v", err)
		c.AddErrorFlash("yaml", "lookup failed")
		return c.NoContent(http.StatusConflict)
	}

	// Load existing YAML data
	yamlDefault, err := loadYamlByViewType(cb, viewType)
	if err != nil {
		log.Printf("loadYamlByViewType failed: %v", err)
		c.AddErrorFlash("yaml", "yaml failed to load")
		return c.NoContent(http.StatusConflict)
	}

	// Process form data
	yamlNoStorage, err := processFormData(formData, unsafeAction)
	if err != nil {
		log.Printf("processFormData failed: %v", err)
		c.AddErrorFlash("yaml", "yaml failed to process")
		return c.NoContent(http.StatusConflict)
	}

	// Update YAML data
	yamlDefault.Cookbook.Pages = yamlNoStorage.Cookbook.Pages
	yamlDefault.Cookbook.Environment = yamlNoStorage.Cookbook.Environment

	// Reorder pages sequentially if needed
	yamlDefault.Cookbook.Pages = models.ReorderPagesSequentially(yamlDefault.Cookbook.Pages)

	// Save YAML back to cookbook
	isAdmin, err := saveYamlToCookbook(cb, yamlDefault, viewType)
	if err != nil {
		log.Printf("saveYamlToCookbook failed: %v", err)
		c.AddErrorFlash("yaml", "yaml failed to save")
		return c.NoContent(http.StatusConflict)
	}

	// Update cookbook in database
	if err := cb.Update(); err != nil {
		log.Printf("Update failed: %v", err)
		c.AddErrorFlash("yaml", "update failed")
		return c.NoContent(http.StatusConflict)
	}

	// Process pages for cache generation
	if err := processPages(yamlDefault, cb, viewType, c.UserContext(), isAdmin); err != nil {
		log.Printf("processPages failed: %v", err)
		c.AddErrorFlash("yaml", "page processing failed")
		return c.NoContent(http.StatusConflict)
	}

	// Build response
	v := models.CoreView{
		Cookbook:    cb,
		ViewType:    viewType,
		YamlDefault: *yamlDefault,
	}

	c.AddSuccessFlash("yaml", "yaml saved")
	re := partials.Cookbook(v)
	return HTML(c, re)
}
