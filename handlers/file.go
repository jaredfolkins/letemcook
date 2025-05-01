package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"gopkg.in/yaml.v3"
)

func LockerDownload(c LemcContext) error {
	log.Println("LockerDownload")
	var username string
	uuid := c.Param("uuid")
	page := c.Param("page")
	scope := strings.ToLower(c.Param("scope"))

	switch scope {
	case SCOPE_YAML_TYPE_SHARED:
		scope = SCOPE_YAML_TYPE_SHARED
		idsa, err := models.GetUserIDsForSharedApp(uuid)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error getting user IDs for shared app %s: %v", uuid, err)
				c.AddErrorFlash("error", "Failed to determine recipients for shared job.")
				return c.NoContent(http.StatusInternalServerError)
			}
		}

		idsc, err := models.GetUserIDsForSharedCookbook(uuid)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error getting user IDs for shared cookbook %s: %v", uuid, err)
				c.AddErrorFlash("error", "Failed to determine recipients for shared job.")
				return c.NoContent(http.StatusInternalServerError)
			}
		}

		ids := append(idsa, idsc...)
		hasAccess := false
		for _, id := range ids {
			if id == c.UserContext().ActingAs.ID {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.AddErrorFlash("error", "You do not have access to this file.")
			return c.NoContent(http.StatusForbidden)
		}

		username = util.SHARED_SINGLETON_USERNAME
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		username = fmt.Sprintf("%s-%v", c.UserContext().ActingAs.Username, c.UserContext().ActingAs.ID)
		scope = SCOPE_YAML_TYPE_INDIVIDUAL

	default:
		c.AddErrorFlash("error", "Invalid scope")
		return c.NoContent(http.StatusBadRequest)
	}

	dir := filepath.Join(os.Getenv("LEMC_LOCKER"), uuid, scope, username, fmt.Sprintf("page-%s", page), "public")
	log.Println("DIRECTORY IS: ", dir)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		log.Printf("Folder exists: %s\n", dir)
	} else {
		log.Printf("Folder does not exist: %s\n", dir)
		c.AddErrorFlash("file-error", "File not found")
		return c.NoContent(http.StatusNotFound)
	}

	c.Response().Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 3600))
	filename := fmt.Sprintf("%s/%s", dir, c.Param("filename"))
	http.ServeFile(c.Response(), c.Request(), filename)
	return nil
}

func GetYamlDownload(c LemcContext) error {
	var yaml_file, userOrAdmin string
	cb := &models.Cookbook{AccountID: 1}
	uuid := strings.TrimSuffix(c.Param("uuid"), ".yaml")
	view_type := c.Param("view_type")

	err := cb.ByUUID(uuid)
	if err != nil {
		log.Println(err)
		return err
	}

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		userOrAdmin = "individual"
		yaml_file, err = prettyPrintYAML(cb.YamlIndividual)
		if err != nil {
			log.Println(err)
			return err
		}
	case SCOPE_YAML_TYPE_SHARED:
		userOrAdmin = "shared"
		yaml_file, err = prettyPrintYAML(cb.YamlShared)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	filename := fmt.Sprintf("%s-%s.yaml", uuid, userOrAdmin)
	tmpFile, err := ioutil.TempFile("", filename)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml_file)); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	c.Response().Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 3600))
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Type", "application/text")
	http.ServeFile(c.Response(), c.Request(), tmpFile.Name())
	return nil
}

func PostYamlUpload(c LemcContext) error {
	uuid := c.Param("uuid")
	view_type := c.Param("view_type")

	file, err := c.FormFile("file")
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

	yaml_default := models.YamlDefault{UUID: cb.UUID}
	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		if err := yaml.Unmarshal(buf.Bytes(), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML")
		}
	case SCOPE_YAML_TYPE_SHARED:
		if err := yaml.Unmarshal(buf.Bytes(), &yaml_default); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to unmarshal yaml YAML")
		}
	default:
		return c.JSON(http.StatusInternalServerError, "Invalid view_type")
	}

	for k, v := range yaml_default.Cookbook.Storage.Wikis {
		// Replace any image links containing the old UUID with the new UUID
		// First, decode the base64 wiki content
		decodedContent, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			continue // Skip this wiki if we can't decode it
		}

		re := regexp.MustCompile(`/wiki/image/([^/]+)/([^/]+)/([^"'\s]+)`)
		updatedContent := re.ReplaceAllString(string(decodedContent), fmt.Sprintf("/wiki/image/$1/%s/$3", uuid))
		yaml_default.Cookbook.Storage.Wikis[k] = base64.StdEncoding.EncodeToString([]byte(updatedContent))
	}

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

	return c.JSON(http.StatusOK, "success")
}
