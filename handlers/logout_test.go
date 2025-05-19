//go:build test

package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// setupTestDB replicates the helper from yeschef tests to create an isolated DB.

func TestPostLogoutHandler(t *testing.T) {
	teardown := db.SetupTestDB(t)
	defer teardown()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/lemc/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	store := sessions.NewCookieStore([]byte("secret"))
	handler := session.Middleware(store)(func(ec echo.Context) error {
		sess, _ := session.Get("session", ec)
		// populate session values to mimic logged in user
		sess.Values["logged_in_user_id"] = int64(1)
		sess.Values["logged_in_account_id"] = int64(1)
		sess.Values["acting_as_user_id"] = int64(1)
		sess.Values["acting_as_account_id"] = int64(1)

		user := &models.User{Account: &models.Account{ID: 1}}
		uc := &models.UserContext{LoggedInAs: user, ActingAs: user}
		cc := middleware.SetUserContext(ec, uc)
		return PostLogoutHandler(cc)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	squid, name, err := util.SquidAndNameByAccountID(1)
	if err != nil {
		t.Fatalf("squid lookup: %v", err)
	}
	expectedURL := fmt.Sprintf("/lemc/login?squid=%s&account=%s", squid, name)
	if got := rec.Header().Get("HX-Replace-Url"); got != expectedURL {
		t.Errorf("HX-Replace-Url header mismatch: got %s want %s", got, expectedURL)
	}

	flashes := rec.Header().Get(middleware.X_LEMC_FLASH_SUCCESS)
	if !strings.Contains(flashes, "logout") {
		t.Errorf("success flash header missing logout key: %s", flashes)
	}

	var found bool
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "session" {
			found = true
			if cookie.MaxAge != -1 {
				t.Errorf("session cookie MaxAge = %d, want -1", cookie.MaxAge)
			}
		}
	}
	if !found {
		t.Error("session cookie not found in response")
	}
}
