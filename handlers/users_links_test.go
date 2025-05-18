package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

func TestUsersPageLinks(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()
	_, user := createUser(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/lemc/account/users?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	c := middleware.SetUserContext(ctx, uc)

	if err := GetAllUsers(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	body := rec.Body.String()
	userLink := "/lemc/account/user/" + strconv.FormatInt(user.ID, 10)
	if !strings.Contains(body, userLink) {
		t.Errorf("expected user detail link %s in body", userLink)
	}
	if !strings.Contains(body, "hx-post=\"/lemc/account/user/create\"") {
		t.Errorf("expected create user action in body")
	}
}
