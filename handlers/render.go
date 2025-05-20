package handlers

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo/v4"
)

func HTML(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func newLoginView(c LemcContext) (models.LoginView, error) {
	squid := c.QueryParam("squid")
	bv := NewBaseView(c)
	loginView := models.LoginView{
		BaseView: bv,
	}

	account, err := models.AccountBySquid(squid)
	if err != nil || account == nil {
		return loginView, err
	}

	newsquid, name, err := models.SquidAndNameByAccountID(account.ID)
	if err != nil {
		return loginView, err
	}

	loginView.BaseView = NewBaseViewWithSquidAndAccountName(c, newsquid, name)
	return loginView, err
}

func renderLoginView(c LemcContext, dv models.LoginView, partial string) error {
	loginView := pages.Login(dv)
	if partial == "true" {
		return HTML(c, loginView)
	}
	li := pages.LoginIndex(dv, loginView)
	return HTML(c, li)
}

func RenderLoginHandler(c LemcContext) error {
	squid := c.Param("squid")

	// Check if the account exists using the squid, but don't need the account object itself here
	_, err := models.AccountBySquid(squid)
	if err != nil {
		log.Printf("Invalid squid provided '%s': %v", squid, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid squid provided"})
	}

	loginView, err := newLoginView(c)
	if err != nil {
		return err
	}

	return renderLoginView(c, loginView, "false")
}
