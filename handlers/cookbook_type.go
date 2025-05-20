package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/yaml.v3"
)

type PostData struct {
	UUID     string `json:"uuid"`
	PageID   string `json:"pageid"`
	Content  string `json:"content"`
	ViewType string `json:"yaml_type"`
}

func PostCookbookWikiSaveHandler(c LemcContext) error {

	var data PostData
	if err := c.Bind(&data); err != nil {
		return err
	}

	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	p.AllowAttrs("class", "style", "width", "height").OnElements("img")
	p.AllowAttrs("class", "style", "width", "height").OnElements("figure")
	p.AllowAttrs("class", "style", "width", "height").OnElements("figcaption")
	p.AllowAttrs("class", "style", "width", "height").OnElements("div")
	data.Content = p.Sanitize(data.Content)

	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cb.ByUUID(data.UUID)
	if err != nil {
		return c.JSON(http.StatusOK, "error")
	}

	v := models.CoreView{Cookbook: cb}
	switch data.ViewType {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
	case SCOPE_YAML_TYPE_SHARED:
		yaml.Unmarshal([]byte(cb.YamlShared), &v.YamlDefault)
	default:
		return fmt.Errorf("view_type not found")
	}

	pageid, err := strconv.Atoi(data.PageID)
	if err != nil {
		return err
	}

	if v.YamlDefault.Cookbook.Storage.Wikis == nil {
		v.YamlDefault.Cookbook.Storage.Wikis = make(map[int]string)
	}
	v.YamlDefault.Cookbook.Storage.Wikis[pageid] = base64.StdEncoding.EncodeToString([]byte(data.Content))

	err = v.YamlDefault.Cookbook.Storage.PurgeUnusedFiles()
	if err != nil {
		return err
	}

	prettyYAML, err := yaml.Marshal(v.YamlDefault)
	if err != nil {
		return err
	}

	switch data.ViewType {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		cb.YamlIndividual = string(prettyYAML)
	case SCOPE_YAML_TYPE_SHARED:
		cb.YamlShared = string(prettyYAML)
	default:
		return fmt.Errorf("view_type not found")
	}
	err = cb.Update()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "success")
}

func GetCookbookEditView(c LemcContext) error {
	uuid := c.Param("uuid")
	view_type := "individual"
	partial := strings.ToLower(c.QueryParam("partial"))

	cb := &models.Cookbook{
		AccountID: c.UserContext().ActingAs.Account.ID,
		UserPerms: &models.PermCookbook{},
	}

	err := cb.ByUUID(uuid)
	if err != nil {
		return err
	}

	err = cb.UserPerms.CookbookPermissions(c.UserContext().ActingAs.ID, c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	var isAdmin bool
	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
		v.YamlDefault.UUID = cb.UUID
		v.ViewType = view_type
		err := preparePages(c, &v, cb.UUID, isAdmin)
		if err != nil {
			return err
		}
	case SCOPE_YAML_TYPE_SHARED:
		yaml.Unmarshal([]byte(cb.YamlShared), &v.YamlDefault)
		v.YamlDefault.UUID = cb.UUID
		v.ViewType = view_type
		isAdmin = true
		err := preparePages(c, &v, cb.UUID, isAdmin)
		if err != nil {
			return err
		}
	case "acls":
		yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
		cba, err := models.CookbookAclsUsers(c.UserContext().ActingAs.Account.ID, cb.ID)
		if err != nil {
			return err
		}

		log.Printf("cba: %v", cba)
		v.YamlDefault.UUID = cb.UUID
		v.ViewType = view_type
		v.CookbookAcls = cba
	}

	var cv templ.Component
	if v.Cookbook.UserPerms.CanEdit {
		cv = pages.Cookbook(v)
	} else {
		cv = pages.App(v)
	}

	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AuthorIndex(v, cv)
	return HTML(c, cvv)
}

func preparePages(c LemcContext, v *models.CoreView, uuid string, isAdmin bool) error {
	for _, page := range v.YamlDefault.Cookbook.Pages {
		jm := &util.JobMeta{
			UUID:     v.YamlDefault.UUID,
			PageID:   strconv.Itoa(page.PageID),
			UserID:   strconv.FormatInt(c.UserContext().ActingAs.ID, 10),
			Username: c.UserContext().ActingAs.Username,
		}

		cf, err := util.NewContainerFiles(jm, isAdmin)
		if err != nil {
			return err
		}

		err = cf.OpenFiles()
		if err != nil {
			return err
		}
		defer cf.CloseFiles()

		css, err := cf.Read(cf.Css)
		if err != nil {
			return err
		}
		page.CssCache = fmt.Sprintf("<style id='uuid-%s-pageid-%d-scope-%s-style'>%s</style>", uuid, page.PageID, v.ViewType, css)

		html, err := cf.Read(cf.Html)
		if err != nil {
			return err
		}
		page.HtmlCache = fmt.Sprintf("<div id='uuid-%s-pageid-%d-scope-%s-html' class='page-inner'>%s</div>", uuid, page.PageID, v.ViewType, html)

		js, err := cf.Read(cf.Js)
		if err != nil {
			return err
		}
		page.JsCache = fmt.Sprintf("<script id='uuid-%s-pageid-%d-scope-%s-script'>%s</script>", uuid, page.PageID, v.ViewType, js)

		v.YamlDefault.Cookbook.Pages[page.PageID-1] = page
	}

	return nil
}

func PostCookbookEditHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	uuid := c.Param("uuid")
	view_type := c.Param("view_type")
	partial := strings.ToLower(c.QueryParam("partial"))
	err := cb.ByUUID(uuid)
	if err != nil {
		log.Printf("PostCookbookEditHandler: %v", err)
		c.AddErrorFlash("cookbook", "Cookbook not found")
		return c.NoContent(http.StatusFound)
	}

	v := models.CoreView{}
	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		cb.YamlIndividual, err = prettyPrintYAML(cb.YamlIndividual)
		if err != nil {
			c.AddErrorFlash("cookbook", "unable to pretty print yaml")
			return c.NoContent(http.StatusFound)
		}
		v.ViewType = "individual"
	case SCOPE_YAML_TYPE_SHARED:
		cb.YamlShared, err = prettyPrintYAML(cb.YamlShared)
		if err != nil {
			c.AddErrorFlash("cookbook", "unable to pretty print yaml")
			return c.NoContent(http.StatusFound)
		}
		v.ViewType = "shared"
	}

	err = cb.Update()
	if err != nil {
		c.AddErrorFlash("cookbook", "unable update cookbook")
		return c.NoContent(http.StatusFound)
	}

	c.AddSuccessFlash("cookbook", "cookbook updated")
	v.Cookbook = cb
	cv := pages.Cookbook(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AuthorIndex(v, cv)
	return HTML(c, cvv)
}

func prettyPrintYAML(yamlStr string) (string, error) {
	var data models.YamlDefault

	err := yaml.Unmarshal([]byte(yamlStr), &data)
	if err != nil {
		return "", err
	}

	prettyYAML, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(prettyYAML), nil
}

func GetCookbookEditIndividualView(c LemcContext) error {
	uuid := c.Param("uuid")
	view_type := "individual"
	partial := strings.ToLower(c.QueryParam("partial"))

	cb := &models.Cookbook{
		AccountID: c.UserContext().ActingAs.Account.ID,
		UserPerms: &models.PermCookbook{},
	}

	err := cb.ByUUID(uuid)
	if err != nil {
		return err
	}

	err = cb.UserPerms.CookbookPermissions(c.UserContext().ActingAs.ID, c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	var isAdmin bool
	yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
	v.YamlDefault.UUID = cb.UUID
	v.ViewType = view_type
	err = preparePages(c, &v, cb.UUID, isAdmin)
	if err != nil {
		return err
	}

	cv := pages.CookbookCompose(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AuthorIndex(v, cv)
	return HTML(c, cvv)
}

func GetCookbookEditSharedView(c LemcContext) error {
	uuid := c.Param("uuid")
	view_type := "shared"
	partial := strings.ToLower(c.QueryParam("partial"))

	cb := &models.Cookbook{
		AccountID: c.UserContext().ActingAs.Account.ID,
		UserPerms: &models.PermCookbook{},
	}

	err := cb.ByUUID(uuid)
	if err != nil {
		return err
	}

	err = cb.UserPerms.CookbookPermissions(c.UserContext().ActingAs.ID, c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	yaml.Unmarshal([]byte(cb.YamlShared), &v.YamlDefault)
	v.YamlDefault.UUID = cb.UUID
	v.ViewType = view_type
	isAdmin := true
	err = preparePages(c, &v, cb.UUID, isAdmin)
	if err != nil {
		return err
	}

	/*
		if v.Cookbook.UserPerms.CanEdit {
			cv = pages.Cookbook(v)
		} else {
			cv = pages.App(v)
		}
	*/

	cv := pages.CookbookCompose(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AuthorIndex(v, cv)
	return HTML(c, cvv)
}

func GetCookbookEditAclsView(c LemcContext) error {
	uuid := c.Param("uuid")
	view_type := "acls"
	partial := strings.ToLower(c.QueryParam("partial"))

	cb := &models.Cookbook{
		AccountID: c.UserContext().ActingAs.Account.ID,
		UserPerms: &models.PermCookbook{},
	}

	err := cb.ByUUID(uuid)
	if err != nil {
		return err
	}

	err = cb.UserPerms.CookbookPermissions(c.UserContext().ActingAs.ID, c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
	cba, err := models.CookbookAclsUsers(c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	v.YamlDefault.UUID = cb.UUID
	v.ViewType = view_type
	v.CookbookAcls = cba

	cv := pages.CookbookAcls(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AuthorIndex(v, cv)
	return HTML(c, cvv)
}
