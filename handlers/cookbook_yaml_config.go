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

func PostCookbookYamlConfigHandler(c LemcContext) error {
	var yaml_default_no_storage models.YamlDefaultNoStorage
	var yaml_default models.YamlDefault
	unsafe_uuid := c.Param("uuid")
	unsafe_action := c.Param("action")
	view_type := c.Param("view_type")

	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	f := c.FormValue(fmt.Sprintf("yaml-%s", unsafe_action))

	err := cb.ByUUID(unsafe_uuid)
	if err != nil {
		log.Printf("ByUUID failed: ", err)
		c.AddErrorFlash("yaml", "lookup failed")
		return c.NoContent(http.StatusConflict)
	}

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		err = yaml.Unmarshal([]byte(cb.YamlIndividual), &yaml_default)
		if err != nil {
			log.Printf("yaml.Unmarshal: ", err)
			c.AddErrorFlash("yaml", "yaml failed to marshal")
			return c.NoContent(http.StatusConflict)
		}
	case SCOPE_YAML_TYPE_SHARED:
		err = yaml.Unmarshal([]byte(cb.YamlShared), &yaml_default)
		if err != nil {
			log.Printf("yaml.Unmarshal: ", err)
			c.AddErrorFlash("yaml", "yaml failed to marshal")
			return c.NoContent(http.StatusConflict)
		}
	default:
		c.AddErrorFlash("yaml", "view_type not found")
		return c.NoContent(http.StatusConflict)
	}

	switch unsafe_action {
	case "all":
		err := yaml.Unmarshal([]byte(f), &yaml_default_no_storage)
		if err != nil {
			c.AddErrorFlash("yaml", "yaml failed to marshal")
			return c.NoContent(http.StatusConflict)
		}
	}

	yaml_default.Cookbook.Pages = yaml_default_no_storage.Cookbook.Pages
	yaml_default.Cookbook.Environment = yaml_default_no_storage.Cookbook.Environment

	prettyYAML, err := yaml.Marshal(yaml_default)
	if err != nil {
		c.AddErrorFlash("yaml", "yaml failed to marshal")
		return c.NoContent(http.StatusConflict)
	}

	var isAdmin bool
	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		cb.YamlIndividual, err = prettyPrintYAML(string(prettyYAML))
		if err != nil {
			c.AddErrorFlash("yaml", "yaml failed to marshal")
			return c.NoContent(http.StatusConflict)
		}
	case SCOPE_YAML_TYPE_SHARED:
		cb.YamlShared, err = prettyPrintYAML(string(prettyYAML))
		if err != nil {
			c.AddErrorFlash("yaml", "yaml failed to marshal")
			return c.NoContent(http.StatusConflict)
		}

		isAdmin = true
	default:
		c.AddErrorFlash("yaml", "view_type not found")
		return c.NoContent(http.StatusConflict)
	}

	err = cb.Update()
	if err != nil {
		log.Printf("Update failed: ", err)
		c.AddErrorFlash("yaml", "update failed")
		return c.NoContent(http.StatusConflict)
	}

	yaml_default.UUID = cb.UUID

	for _, page := range yaml_default.Cookbook.Pages {

		jm := &util.JobMeta{
			UUID:     cb.UUID,
			PageID:   strconv.Itoa(page.PageID),
			UserID:   fmt.Sprintf("%d", c.UserContext().ActingAs.Account.ID),
			Username: c.UserContext().ActingAs.Username,
		}

		cf, err := util.NewContainerFiles(jm, isAdmin)
		if err != nil {
			log.Printf("NewContainerFiles failed: ", err)
			c.AddErrorFlash("container_files", "container files failed")
			return c.NoContent(http.StatusConflict)
		}

		err = cf.OpenFiles()
		if err != nil {
			log.Printf("OpenFiles failed: ", err)
			c.AddErrorFlash("container_files", "open files failed")
			return c.NoContent(http.StatusConflict)
		}
		defer cf.CloseFiles()

		css, err := cf.Read(cf.Css)
		if err != nil {
			log.Printf("Read failed: ", err)
			c.AddErrorFlash("read", "reading files failed")
			return c.NoContent(http.StatusConflict)
		}
		page.CssCache = fmt.Sprintf("<style id='uuid-%s-pageid-%d-scope-%s-style'>%s</style>", cb.UUID, page.PageID, view_type, css)

		html, err := cf.Read(cf.Html)
		if err != nil {
			log.Printf("Read failed: ", err)
			c.AddErrorFlash("read", "reading files failed")
			return c.NoContent(http.StatusConflict)
		}
		page.HtmlCache = fmt.Sprintf("<div id='uuid-%s-pageid-%d-scope-%s-html' class='page-inner'>%s</div>", cb.UUID, page.PageID, view_type, html)

		js, err := cf.Read(cf.Js)
		if err != nil {
			log.Printf("Read failed: ", err)
			c.AddErrorFlash("read", "reading files failed")
			return c.NoContent(http.StatusConflict)
		}
		page.JsCache = fmt.Sprintf("<script id='uuid-%s-pageid-%d-scope-%s-script'>%s</script>", cb.UUID, page.PageID, view_type, js)

		yaml_default_no_storage.Cookbook.Pages[page.PageID-1] = page
	}

	v := models.CoreView{Cookbook: cb, ViewType: view_type}
	v.YamlDefault = yaml_default

	c.AddSuccessFlash("yaml", "yaml saved")
	re := partials.Cookbook(v)
	return HTML(c, re)
}
