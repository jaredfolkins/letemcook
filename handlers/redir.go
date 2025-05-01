package handlers

import (
	"fmt"
	"net/http"

	"github.com/jaredfolkins/letemcook/util"
)

func redirLoginHandler(c LemcContext) error {
	id, name, err := util.SquidAndNameByAccountID(1)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/setup")
	}

	redir := fmt.Sprintf("/lemc/login?squid=%s&account=%s", id, name)
	return c.Redirect(http.StatusTemporaryRedirect, redir)
}

func redirRegisterHandler(c LemcContext) error {
	id, name, err := util.SquidAndNameByAccountID(1)
	if err != nil {
		return err
	}
	redir := fmt.Sprintf("/lemc/register?squid=%s&account=%s", id, name)
	return c.Redirect(http.StatusTemporaryRedirect, redir)
}
