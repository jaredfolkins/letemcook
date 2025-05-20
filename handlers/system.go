package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/jaredfolkins/letemcook/yeschef"
)

func getSystemView(c LemcContext) models.SystemView {
	bv := NewBaseView(c)
	bv.ActiveNav = "system"
	return models.SystemView{BaseView: bv}
}

func GetSystemSettingsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemSettings

	settings := map[string]string{
		"LEMC_ENV":         os.Getenv("LEMC_ENV"),
		"LEMC_FQDN":        os.Getenv("LEMC_FQDN"),
		"LEMC_DATA":        os.Getenv("LEMC_DATA"),
		"LEMC_PORT_DEV":    os.Getenv("LEMC_PORT_DEV"),
		"LEMC_PORT_TEST":   os.Getenv("LEMC_PORT_TEST"),
		"LEMC_PORT_PROD":   os.Getenv("LEMC_PORT_PROD"),
		"LEMC_DOCKER_HOST": os.Getenv("LEMC_DOCKER_HOST"),
	}
	sv := models.SystemSettingsView{BaseView: v.BaseView, Settings: settings}
	cmp := pages.SystemSettings(sv)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemSettingsIndex(sv, cmp))
}

func GetSystemAccountsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemAccounts

	pageParam := c.QueryParam("page")
	limitParam := c.QueryParam("limit")
	page, _ := strconv.Atoi(pageParam)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitParam)
	if limit < 1 {
		limit = 10
	}
	accounts, err := models.Accounts(page, limit)
	if err != nil {
		return err
	}
	total, err := models.CountAccounts()
	if err != nil {
		return err
	}
	totalPages := (total + limit - 1) / limit
	sv := models.SystemAccountsView{BaseView: v.BaseView, Accounts: accounts, CurrentPage: page, TotalPages: totalPages, Limit: limit}
	cmp := pages.SystemAccounts(sv)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemAccountsIndex(sv, cmp))
}

func GetSystemImagesHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemImages

	imgs, err := models.CollectImageInfos()
	if err != nil {
		return err
	}
	sv := models.SystemImagesView{BaseView: v.BaseView, Images: imgs}
	cmp := pages.SystemImages(sv)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemImagesIndex(sv, cmp))
}

func GetSystemJobsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemJobs

	pageParam := c.QueryParam("page")
	limitParam := c.QueryParam("limit")
	page, _ := strconv.Atoi(pageParam)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitParam)
	if limit < 1 {
		limit = 10
	}
	jobs, total, err := getAllJobs(page, limit)
	if err != nil {
		return err
	}
	totalPages := (total + limit - 1) / limit
	sv := models.SystemJobsView{BaseView: v.BaseView, Jobs: jobs, CurrentPage: page, TotalPages: totalPages, Limit: limit}
	cmp := pages.SystemJobs(sv)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemJobsIndex(sv, cmp))
}

func PostSystemImagePullHandler(c LemcContext) error {
	img := c.FormValue("image")
	if img == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := yeschef.PullImage(yeschef.ImageSpec{Name: img}); err != nil {
		return err
	}

	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemImages
	imgs, err := models.CollectImageInfos()
	if err != nil {
		return err
	}
	sv := models.SystemImagesView{BaseView: v.BaseView, Images: imgs}
	cmp := pages.SystemImages(sv)
	return HTML(c, cmp)
}
