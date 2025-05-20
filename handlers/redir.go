package handlers

import (
	"fmt"
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
)

func redirLoginHandler(c LemcContext) error {
	id, name, err := models.SquidAndNameByAccountID(1)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/setup")
	}

	redir := fmt.Sprintf("%s?squid=%s&account=%s", paths.Login, id, name)
	return c.Redirect(http.StatusTemporaryRedirect, redir)
}

func redirRegisterHandler(c LemcContext) error {
	id, name, err := models.SquidAndNameByAccountID(1)
	if err != nil {
		return err
	}
	redir := fmt.Sprintf("%s?squid=%s&account=%s", paths.Register, id, name)
	return c.Redirect(http.StatusTemporaryRedirect, redir)
}
