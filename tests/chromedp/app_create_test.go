package main_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestCreateApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := createHeadlessContext(t)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	loginVals := url.Values{}
	loginVals.Set("squid", "xkQN")
	loginVals.Set("account", "Account Alpha")
	loginURL := baseURL + loginPath + "?" + loginVals.Encode()

	appName := fmt.Sprintf("TestApp-%d", time.Now().UnixNano())
	appDesc := "created during chromedp test"
	cookbookSearch := "Alpha Cookbook 1"

	var successFlash string

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(usernameSelector, validUsername, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, validPassword, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),

		chromedp.Navigate(baseURL + "/lemc/apps"),
		chromedp.WaitVisible(`button[onclick="new_app.showModal()"]`, chromedp.ByQuery),
		chromedp.Click(`button[onclick="new_app.showModal()"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`#new_app`, chromedp.ByID),
		chromedp.SendKeys(`#new-app-form input[name="name"]`, appName, chromedp.ByQuery),
		chromedp.SendKeys(`#new-app-form textarea[name="description"]`, appDesc, chromedp.ByQuery),
		chromedp.SendKeys(`#acl-search-input`, cookbookSearch, chromedp.ByID),
		chromedp.WaitVisible(`#acl-search-display`, chromedp.ByID),
		chromedp.Click(`#acl-search-display div`, chromedp.ByQuery),
		chromedp.Click(`#new-app-form button`, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),
		chromedp.Text(flashSuccessSelector, &successFlash, chromedp.ByQuery),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed running chromedp tasks: %v", err)
	}

	if !strings.Contains(successFlash, "new app created") {
		t.Errorf("expected success flash to contain 'new app created', got %q", successFlash)
	}
}
