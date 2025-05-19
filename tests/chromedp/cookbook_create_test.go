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

func TestCreateCookbook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := createHeadlessContext(t)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	// Build login URL
	loginVals := url.Values{}
	loginVals.Set("squid", "xkQN")
	loginVals.Set("account", "Account Alpha")
	loginURL := baseURL + loginPath + "?" + loginVals.Encode()

	cookbookName := fmt.Sprintf("TestCookbook-%d", time.Now().UnixNano())
	cookbookDesc := "created during chromedp test"

	var successFlash string

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(usernameSelector, validUsername, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, validPassword, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),

		chromedp.Navigate(baseURL + "/lemc/cookbooks"),
		chromedp.WaitVisible(`button[onclick="new_cookbook.showModal()"]`, chromedp.ByQuery),
		chromedp.Click(`button[onclick="new_cookbook.showModal()"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`#new_cookbook`, chromedp.ByID),
		chromedp.SendKeys(`#new-cookbook-form input[name="name"]`, cookbookName, chromedp.ByQuery),
		chromedp.SendKeys(`#new-cookbook-form textarea[name="description"]`, cookbookDesc, chromedp.ByQuery),
		chromedp.Click(`#new-cookbook-form button`, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),
		chromedp.Text(flashSuccessSelector, &successFlash, chromedp.ByQuery),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed running chromedp tasks: %v", err)
	}

	if !strings.Contains(successFlash, "new cookbook created") {
		t.Errorf("expected success flash to contain 'new cookbook created', got %q", successFlash)
	}
}
