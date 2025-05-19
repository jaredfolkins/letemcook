package handlers

import (
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/views/pages"
)

func getSystemView(c LemcContext) models.SystemView {
	bv := NewBaseView(c)
	bv.ActiveNav = "system"
	return models.SystemView{BaseView: bv}
}

func GetSystemSettingsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemSettings
	cmp := pages.SystemSettings(v)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemSettingsIndex(v, cmp))
}

func GetSystemAccountsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemAccounts
	cmp := pages.SystemAccounts(v)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemAccountsIndex(v, cmp))
}

func GetSystemImagesHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemImages
	cmp := pages.SystemImages(v)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemImagesIndex(v, cmp))
}

func GetSystemJobsHandler(c LemcContext) error {
	v := getSystemView(c)
	v.BaseView.ActiveSubNav = paths.SystemJobs
	cmp := pages.SystemJobs(v)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.SystemJobsIndex(v, cmp))
}
