package handlers

import (
	"fmt"
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo-contrib/session"
)

func PostLogoutHandler(c LemcContext) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	newSquid, newName, err := util.SquidAndNameByAccountID(c.UserContext().LoggedInAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	// We can't directly modify the userContext, but we can clear the session
	// The Before middleware will handle the user context

	delete(sess.Values, "logged_in_user_id")
	delete(sess.Values, "logged_in_account_id")
	delete(sess.Values, "acting_as_user_id")
	delete(sess.Values, "acting_as_account_id")
	sess.Options.MaxAge = -1 // Expire the cookie immediately

	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	c.AddSuccessFlash("logout", "You have logged out")

	loginView := models.LoginView{
		BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	c.Response().Header().Set("HX-Replace-Url", fmt.Sprintf("/lemc/login?squid=%s&account=%s", newSquid, newName))
	loginPage := pages.Login(loginView)
	return HTML(c, loginPage)
}
