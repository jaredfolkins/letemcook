package handlers

import (
	"log"

	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/partials"

	"github.com/jaredfolkins/letemcook/models"
)

func GetNavtop(c LemcContext) error {
	nsquid := "squid-not-found"
	nname := "name-not-found"
	squid := c.QueryParam("squid")
	account, err := models.AccountBySquid(squid)
	if err != nil {
		log.Printf("Failed to find account by squid '%s': %v", squid, err)
	} else {
		nsquid, nname, _ = util.SquidAndNameByAccountID(account.ID)
	}

	section := c.QueryParam("section")
	subnav := c.QueryParam("subnav")
	bv := NewBaseViewWithSquidAndAccountName(c, nsquid, nname)
	bv.ActiveNav = section
	bv.ActiveSubNav = subnav
	nt := partials.Navtop(bv)
	return HTML(c, nt)
}
