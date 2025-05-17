package handlers

import (
	"log"

	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/partials"

	"github.com/jaredfolkins/letemcook/models"
)

func GetNavtop(c LemcContext) error {
	squid := c.QueryParam("squid")

	var bv models.BaseView

	if squid != "" {
		if account, err := models.AccountBySquid(squid); err == nil {
			nsquid, nname, _ := util.SquidAndNameByAccountID(account.ID)
			bv = NewBaseViewWithSquidAndAccountName(c, nsquid, nname)
		} else {
			bv = NewBaseView(c)
		}
	} else {
		bv = NewBaseView(c)
	}

	nt := partials.Navtop(bv)
	return HTML(c, nt)
}
