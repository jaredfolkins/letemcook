package handlers

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
)

const DefaultCookbookLimit = 10

func GetCookbooksHandler(c LemcContext) error {

	partial := strings.ToLower(c.QueryParam("partial"))
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = DefaultCookbookLimit
	}

	userID := c.UserContext().ActingAs.ID

	totalCookbooks, err := models.CountCookbooks(userID)
	if err != nil {
		log.Printf("Error counting cookbooks: %v", err)
		return err
	}
	totalPages := 0
	if totalCookbooks > 0 {
		totalPages = int(math.Ceil(float64(totalCookbooks) / float64(limit)))
	}

	cbs, err := models.Cookbooks(userID, page, limit)
	if err != nil {
		log.Printf("Error fetching cookbooks: %v", err)
		return err
	}

	newSquid, newName, err := util.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.CookbooksView{
		Cookbooks:   cbs,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
		BaseView:    NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}
	v.BaseView.ActiveNav = "cookbooks"

	cv := pages.Cookbooks(v)
	if partial == "true" {
		return HTML(c, cv)
	}

	cvv := pages.CookbooksIndex(v, cv)
	return HTML(c, cvv)
}
