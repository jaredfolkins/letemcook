package handlers

import (
	"log"
	"strings"

	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/views/partials"

	"github.com/jaredfolkins/letemcook/models"
)

func GetNavtop(c LemcContext) error {
	log.Printf("GetNavtop: Handler called")

	nsquid := "squid-not-found"
	nname := "name-not-found"
	squid := c.QueryParam("squid")
	account, err := models.AccountBySquid(squid)
	if err != nil {
		log.Printf("Failed to find account by squid '%s': %v", squid, err)
	} else {
		nsquid, nname, _ = models.SquidAndNameByAccountID(account.ID)
	}

	section := c.QueryParam("section")
	subnav := c.QueryParam("subnav")

	log.Printf("GetNavtop: section='%s', subnav='%s'", section, subnav)

	// If section is empty, try to detect from referer URL
	if section == "" || section == "undefined" {
		referer := c.Request().Header.Get("Referer")
		log.Printf("GetNavtop: Section empty, checking referer: '%s'", referer)
		log.Printf("GetNavtop: paths.Apps is: '%s'", paths.Apps)
		if referer != "" {
			if strings.Contains(referer, paths.Apps) {
				section = "apps"
				log.Printf("GetNavtop: Detected section 'apps' from referer")
			} else if strings.Contains(referer, paths.Cookbooks) {
				section = "cookbooks"
				log.Printf("GetNavtop: Detected section 'cookbooks' from referer")
			} else if strings.Contains(referer, paths.AccountSettings) || strings.Contains(referer, paths.AccountUsers) || strings.Contains(referer, paths.AccountJobs) {
				section = "account"
				log.Printf("GetNavtop: Detected section 'account' from referer")
			} else if strings.Contains(referer, paths.SystemSettings) || strings.Contains(referer, paths.SystemAccounts) || strings.Contains(referer, paths.SystemImages) || strings.Contains(referer, paths.SystemJobs) {
				section = "system"
				log.Printf("GetNavtop: Detected section 'system' from referer")
			}
		}
	}

	log.Printf("GetNavtop: Final section='%s', subnav='%s'", section, subnav)

	bv := NewBaseViewWithSquidAndAccountName(c, nsquid, nname)
	bv.ActiveNav = section
	bv.ActiveSubNav = subnav
	nt := partials.Navtop(bv)
	return HTML(c, nt)
}
