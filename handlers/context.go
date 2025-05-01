package handlers

import (
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/labstack/echo/v4"
)

type Handler func(middleware.LemcContext) error

func Ctx(h Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(c.(middleware.LemcContext))
	}
}
