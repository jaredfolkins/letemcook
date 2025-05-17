package handlers

import (
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/errors"
	"github.com/labstack/echo/v4"

	"github.com/a-h/templ"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	c.Logger().Error(err)

	var errorPage func(v models.BaseView) templ.Component

	switch code {
	case http.StatusUnauthorized: // 401
		errorPage = errors.Error401
	case http.StatusForbidden: // 403
		errorPage = errors.Error403
	case http.StatusNotFound: // 404
		errorPage = errors.Error404
	case http.StatusMethodNotAllowed: // 405
		errorPage = errors.Error405
	case http.StatusInternalServerError: // 500
		fallthrough // Fall through to default for 500
	default:
		code = http.StatusInternalServerError // Ensure code is 500 for default
		errorPage = errors.Error500
	}

	c.Response().WriteHeader(code)
	if errorPage != nil {
		HTML(c, errorPage(models.BaseView{ActiveNav: ""}))
	} else {
		c.Logger().Errorf("Error page handler is nil for code %d", code)
		_ = c.String(code, http.StatusText(code))
	}
}
