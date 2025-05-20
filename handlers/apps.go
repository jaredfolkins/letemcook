package handlers

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo/v4"
)

const DefaultappLimit = 10

func GetAppsHandler(c LemcContext) error {

	partial := strings.ToLower(c.QueryParam("partial"))
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = DefaultappLimit
	}

	if !c.UserContext().IsAuthenticated() || c.UserContext().ActingAs == nil || c.UserContext().ActingAs.Account == nil {
		log.Printf("GetAppsHandler: Unauthenticated or incomplete user context detected.")
		return echo.NewHTTPError(http.StatusUnauthorized, "User context not available")
	}

	userID := c.UserContext().ActingAs.ID
	accountID := c.UserContext().ActingAs.Account.ID

	totalapps, err := models.Countapps(userID, accountID)
	if err != nil {
		log.Printf("Error counting apps: %v", err)
		return err
	}
	totalPages := 0
	if totalapps > 0 {
		totalPages = int(math.Ceil(float64(totalapps) / float64(limit)))
	}

	apps, err := models.Apps(userID, accountID, page, limit)
	if err != nil {
		log.Printf("Error fetching apps: %v", err)
		return err
	}

	if c.UserContext().ActingAs == nil || c.UserContext().ActingAs.Account == nil {
		log.Printf("GetAppsHandler: ActingAs context became nil unexpectedly before fetching Squid.")
		return echo.NewHTTPError(http.StatusInternalServerError, "User context lost")
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.AppsView{
		Apps:        apps,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
		BaseView:    NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}
	v.BaseView.ActiveNav = "apps"

	cv := pages.Apps(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.AppsIndex(v, cv)
	return HTML(c, cvv)
}
