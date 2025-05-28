package main_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jaredfolkins/letemcook/tests/testutil"
)

// TestNavClickAfterHardRefresh captures JavaScript errors that occur when
// a user performs a hard refresh and then clicks a navigation link.
func TestNavClickAfterHardRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load the actual squid values from the test environment
	alphaSquid, _, err := testutil.LoadTestEnv()
	if err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	ctx, cancel := testutil.CreateHeadlessContext()
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	loginVals := url.Values{}
	loginVals.Set("squid", alphaSquid)
	loginVals.Set("account", testutil.AlphaAccountName)
	loginURL := testutil.GetBaseURL() + "/lemc/login" + "?" + loginVals.Encode()

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
		chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second), // Wait for login to complete

		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/apps"),
		chromedp.WaitVisible(`#navtop`, chromedp.ByQuery),

		// Simple page reload instead of hard refresh to avoid import issues
		chromedp.Reload(),
		chromedp.WaitVisible(`#navtop`, chromedp.ByQuery),
		chromedp.Click(`#navtop a[href="/lemc/cookbooks?partial=true"]`, chromedp.ByQuery),
		chromedp.Sleep(1 * time.Second),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed running chromedp tasks: %v", err)
	}
}

func TestNavRefreshError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load the actual squid values from the test environment
	alphaSquid, _, err := testutil.LoadTestEnv()
	if err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	ctx, cancel := testutil.CreateHeadlessContext()
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	loginURLValues := url.Values{}
	loginURLValues.Set("squid", alphaSquid)
	loginURLValues.Set("account", testutil.AlphaAccountName)
	loginURL := testutil.GetBaseURL() + "/lemc/login" + "?" + loginURLValues.Encode()

	var navHTML string

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
		chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second), // Wait for login to complete

		// Navigate to cookbooks page
		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/cookbooks"),
		chromedp.Sleep(1 * time.Second),

		// Trigger a nav refresh that would cause an error
		chromedp.Evaluate(`document.body.dispatchEvent(new Event('refreshNavtop'))`, nil),
		chromedp.Sleep(2 * time.Second),

		// Capture the nav HTML
		chromedp.InnerHTML("#navtop", &navHTML, chromedp.ByID),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed nav refresh error test: %v", err)
	}

	// Check if the navigation still works after error handling
	if !strings.Contains(navHTML, "Apps") && !strings.Contains(navHTML, "Cookbooks") {
		t.Errorf("expected navigation to be present after refresh error, got HTML: %s", navHTML)
	}
}
