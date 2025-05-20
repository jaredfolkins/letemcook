package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/jaredfolkins/letemcook/views/partials"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

func GetCookbookSearchByName(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseView(c),
	}

	search := c.FormValue("cookbook-search")
	if len(search) < 1 {
		dsr := partials.DisplayAclSearchResults(v)
		return HTML(c, dsr)
	}
	is_published := true
	cba, err := models.SearchForCookbooks(search, c.UserContext().ActingAs.ID, c.UserContext().ActingAs.Account.ID, 100, is_published)
	if err != nil {
		log.Println("SearchForCookbooks: ", err)
	}

	for i, vcba := range cba {
		cba[i].HtmlName = highlightMatch(vcba.Name, search)
		cba[i].HtmlDescription = highlightMatch(vcba.Description, search)
	}

	v.CookbookSearchResults = cba
	dsr := partials.DisplayCookbookSearchResults(v)
	return HTML(c, dsr)
}

func PostCookbookCreate(c LemcContext) error {
	cb := &models.Cookbook{
		AccountID:   c.UserContext().ActingAs.Account.ID,
		OwnerID:     c.UserContext().ActingAs.ID,
		Name:        util.Sanitize(c.FormValue("name")),
		Description: util.Sanitize(c.FormValue("description")),
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	err = cb.ByName(cb.Name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("error creating cookbook: ", err)
		c.AddErrorFlash("cookbook-create", "sql ByName() error")
		return c.NoContent(http.StatusConflict)
	}

	if cb.ID > 0 {
		c.AddErrorFlash("cookbook-create", fmt.Sprintf("cookbook [%s] already exists", cb.Name))
		return c.NoContent(http.StatusConflict)
	}

	err = cb.Create(tx)
	if err != nil {
		log.Println("error creating cookbook: ", err)
		errR := tx.Rollback()
		if errR != nil {
			log.Println("🔥 Failed to rollback transaction: ", errR)
		}
		c.AddErrorFlash("cookbook-create", "server error creating cookbook")
		return c.NoContent(http.StatusConflict)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	userID := c.UserContext().ActingAs.ID

	limit := DefaultCookbookLimit // Use the defined constant

	totalCookbooks, err := models.CountCookbooks(userID)
	if err != nil {
		log.Printf("Error counting cookbooks after create: %v", err)
	}

	totalPages := 0
	if totalCookbooks > 0 {
		totalPages = int(math.Ceil(float64(totalCookbooks) / float64(limit)))
	}

	page := 1

	cbs, err := models.Cookbooks(userID, page, limit) // Use calculated page and limit
	if err != nil {
		log.Printf("Error fetching cookbooks after create: %v", err)
		return err // Or return a specific error response
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CookbooksView{
		Cookbooks:   cbs,
		CurrentPage: page,       // Add CurrentPage
		TotalPages:  totalPages, // Add TotalPages
		Limit:       limit,      // Add Limit
		BaseView:    NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}
	v.BaseView.ActiveNav = "cookbooks"

	c.AddSuccessFlash("cookbook-create", "new cookbook created")
	cv := pages.CookbooksList(v)

	c.Response().Header().Set("HX-Trigger", "closeNewCookbookModal")

	pushedURL := fmt.Sprintf("%s?page=1&limit=%d", paths.Cookbooks, limit)
	c.Response().Header().Set("HX-Push-Url", pushedURL)

	return HTML(c, cv)
}

func PostCookbookWikiImage(c echo.Context) error {
	uuid := c.Param("uuid")
	view_type := c.Param("view_type")

	file, err := c.FormFile("upload")
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to retrieve the file")
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to open the file")
	}
	defer src.Close()

	var buf bytes.Buffer

	if _, err = io.Copy(&buf, src); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to copy the file to buffer")
	}

	cb := &models.Cookbook{}
	if err := cb.ByUUID(uuid); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to to open Cookbook")
	}

	yaml_default := models.YamlDefault{}
	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		if err := yaml.Unmarshal([]byte(cb.YamlIndividual), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML")
		}
	case SCOPE_YAML_TYPE_SHARED:
		if err := yaml.Unmarshal([]byte(cb.YamlShared), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal yaml YAML")
		}
	default:
		return c.JSON(http.StatusInternalServerError, "Invalid view_type")
	}

	filename := util.ReplaceSpecialCharsWithDashes(strings.ToLower(file.Filename))
	yaml_default.UUID = cb.UUID
	yaml_default.Cookbook.Storage.Files[filename] = base64.StdEncoding.EncodeToString(buf.Bytes())

	prettyYAML, err := yaml.Marshal(yaml_default)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to marshal YAML")
	}

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		cb.YamlIndividual = string(prettyYAML)
	case SCOPE_YAML_TYPE_SHARED:
		cb.YamlShared = string(prettyYAML)
	default:
		return c.JSON(http.StatusInternalServerError, "Invalid view_type")
	}

	err = cb.Update()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to update Cookbook")
	}

	wikiImage := &Image{
		Url: fmt.Sprintf("/lemc/wiki/image/%s/%s/%s", view_type, uuid, filename),
	}

	return c.JSON(http.StatusOK, wikiImage)
}

func GetCookbookWikiImage(c LemcContext) error {
	var err error
	var decoded []byte
	var yaml_individual, yaml_shared string
	var app *models.App

	uuid := c.Param("uuid")
	filename := c.Param("filename")
	cb := &models.Cookbook{}
	err = cb.ByUUID(uuid)
	if err != nil {
		log.Println("error getting cookbook: ", uuid, err)
	}

	yaml_individual = cb.YamlIndividual
	yaml_shared = cb.YamlShared

	if len(cb.UUID) < 4 && err != nil {
		app, err = models.AppByUUIDAndAccountID(uuid, c.UserContext().ActingAs.Account.ID)
		if err != nil {
			log.Println("error getting app: ", err)
			return c.JSON(http.StatusInternalServerError, "Failed to to open Cookbook or App")
		}

		yaml_individual = app.YAMLIndividual
		yaml_shared = app.YAMLShared
	}

	yaml_default := models.YamlDefault{}
	switch c.Param("view_type") {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		if err := yaml.Unmarshal([]byte(yaml_individual), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML")
		}
	case SCOPE_YAML_TYPE_SHARED:
		if err := yaml.Unmarshal([]byte(yaml_shared), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal yaml YAML")
		}
	default:
		return c.JSON(http.StatusInternalServerError, "Invalid view_type")
	}

	for k, v := range yaml_default.Cookbook.Storage.Files {
		if k == filename {
			decoded, err = base64.StdEncoding.DecodeString(v)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "Failed to decode base64 encoded file")
			}
		}
	}

	return c.Blob(http.StatusOK, http.DetectContentType(decoded), decoded)
}

type Image struct {
	Url string `json:"url"`
}

func PostCookbookMetaUpdateHandler(c LemcContext) error {
	unsafe_uuid := c.Param("uuid")
	desc := util.Sanitize(c.FormValue("cookbook_desc"))
	name := util.Sanitize(c.FormValue("cookbook_name"))

	cbOrig := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cbOrig.ByName(name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("error updating cookbook: ", err)
		c.AddErrorFlash("cookbook-update", "sql lookup failed")
		return c.NoContent(http.StatusConflict)
	}

	if len(cbOrig.Name) > 0 {
		c.AddErrorFlash("cookbook-update", fmt.Sprintf("name must be unique but [%s] is already in use", cbOrig.Name))
		return c.NoContent(http.StatusConflict)
	}

	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err = cb.ByUUID(unsafe_uuid)
	if err != nil {
		log.Println("error updating cookbook: ", err)
		c.AddErrorFlash("cookbook-update", "unable to find cookbook")
		return c.NoContent(http.StatusConflict)
	}

	if len(name) > 0 {
		cb.Name = name
	}

	if len(desc) > 0 {
		cb.Description = desc
	}

	if len(name) == 0 && len(desc) == 0 {
		c.AddErrorFlash("cookbook-update", "no changes detected")
		return c.NoContent(http.StatusConflict)
	}

	err = cb.Update()
	if err != nil {
		log.Println("error updating cookbook: ", err)
		c.AddErrorFlash("cookbook-update", "unable to update name or description")
		return c.NoContent(http.StatusConflict)
	}

	if len(name) > 0 {
		c.AddSuccessFlash("cookbook-update-name", "name updated")
	}

	if len(desc) > 0 {
		c.AddSuccessFlash("cookbook-update-desc", "description updated")
	}

	return c.NoContent(http.StatusOK)
}
