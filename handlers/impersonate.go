package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/jaredfolkins/letemcook/views/partials"
	"github.com/labstack/echo-contrib/session"
)

func GetImpersonateHandler(c LemcContext) error {
	v := models.ImpersonateView{BaseView: NewBaseView(c)}
	v.BaseView.ActiveNav = "system"
	v.BaseView.ActiveSubNav = paths.Impersonate
	cmp := pages.Impersonate(v)
	if strings.ToLower(c.QueryParam("partial")) == "true" {
		return HTML(c, cmp)
	}
	return HTML(c, pages.ImpersonateIndex(v, cmp))
}

func GetImpersonateSearchHandler(c LemcContext) error {
	v := models.ImpersonateView{BaseView: NewBaseView(c)}
	search := c.FormValue("impersonate-search")
	if len(search) > 0 {
		users, err := models.SearchAllUsers(search, 100)
		if err != nil {
			log.Printf("impersonate search error: %v", err)
		}
		v.Users = users
	}
	cmp := partials.DisplayImpersonateSearchResults(v)
	return HTML(c, cmp)
}

func PostImpersonateHandler(c LemcContext) error {
	uidStr := c.Param("uid")
	aidStr := c.Param("aid")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad uid")
	}
	aid, err := strconv.ParseInt(aidStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad aid")
	}
	user, err := models.UserByIDAndAccountID(uid, aid)
	if err != nil {
		return c.String(http.StatusNotFound, "user not found")
	}
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Values["acting_as_user_id"] = user.ID
	sess.Values["acting_as_account_id"] = user.Account.ID
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	c.AddSuccessFlash("impersonate", fmt.Sprintf("Now impersonating %s", user.Username))
	return c.NoContent(http.StatusOK)
}
