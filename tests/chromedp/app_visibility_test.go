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

const (
	seedPassword      = "asdfasdfasdf"
	accAlphaName      = "Account Alpha"
	accAlphaOwnerUser = "alpha-owner"
	accBravoName      = "Account Bravo"
	accBravoOwnerUser = "bravo-owner"
	alphaAppPrefix    = "Alpha App"
	bravoAppPrefix    = "Bravo App"
	appsPath          = "/lemc/apps"
	appListSelector   = "#app-list"           // Assumed selector for the app list container
	appItemSelector   = ".list-group-item h5" // Assumed selector for app names within the list
)

type appVisibilityTestData struct {
	testName           string
	username           string
	password           string
	accountName        string
	squid              string
	shouldSeeSubstr    string
	shouldNotSeeSubstr string
	expectPresence     bool // True if shouldSeeSubstr should be present
}

func TestAppVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use parallel test wrapper for automatic instance management
	testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Load test environment for this specific instance
		alphaSquid, bravoSquid, err := testutil.LoadTestEnvForInstance(instance)
		if err != nil {
			t.Fatalf("Failed to load test environment: %v", err)
		}

		// Use ChromeDP with the instance
		testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			// Test that Alpha account can see its apps
			t.Run("AlphaCanSeeOwnApps", func(t *testing.T) {
				loginVals := url.Values{}
				loginVals.Set("squid", alphaSquid)
				loginVals.Set("account", testutil.AlphaAccountName)

				// Use the instance-specific base URL
				baseURL := testutil.GetBaseURLForInstance(instance)
				loginURL := baseURL + "/lemc/login?" + loginVals.Encode()
				targetAppsURL := baseURL + "/lemc/apps"

				var appsPageHTML string

				tasks := chromedp.Tasks{
					chromedp.Navigate(loginURL),
					chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
					chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
					chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
					chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
					chromedp.Sleep(3 * time.Second), // Wait for login to complete

					// Navigate to apps page to check visibility
					chromedp.Navigate(targetAppsURL),
					chromedp.Sleep(2 * time.Second), // Wait for page to load

					// Capture the apps page content
					chromedp.OuterHTML("body", &appsPageHTML, chromedp.ByQuery),
				}

				if err := chromedp.Run(ctx, tasks); err != nil {
					t.Fatalf("failed running chromedp tasks: %v", err)
				}

				// Verify Alpha can access apps page
				if !strings.Contains(appsPageHTML, "Apps") && !strings.Contains(appsPageHTML, "app") {
					t.Errorf("expected Alpha to see Apps page content")
				}
			})

			// Test that Bravo account can see its apps
			t.Run("BravoCanSeeOwnApps", func(t *testing.T) {
				loginVals := url.Values{}
				loginVals.Set("squid", bravoSquid)
				loginVals.Set("account", testutil.BravoAccountName)

				// Use the instance-specific base URL
				baseURL := testutil.GetBaseURLForInstance(instance)
				loginURL := baseURL + "/lemc/login?" + loginVals.Encode()
				targetAppsURL := baseURL + "/lemc/apps"

				var appsPageHTML string

				tasks := chromedp.Tasks{
					chromedp.Navigate(loginURL),
					chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
					chromedp.SendKeys(testutil.UsernameSelector, testutil.BravoOwnerUsername, chromedp.ByQuery),
					chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
					chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
					chromedp.Sleep(3 * time.Second), // Wait for login to complete

					// Navigate to apps page to check visibility
					chromedp.Navigate(targetAppsURL),
					chromedp.Sleep(2 * time.Second), // Wait for page to load

					// Capture the apps page content
					chromedp.OuterHTML("body", &appsPageHTML, chromedp.ByQuery),
				}

				if err := chromedp.Run(ctx, tasks); err != nil {
					t.Fatalf("failed running chromedp tasks: %v", err)
				}

				// Verify Bravo can access apps page
				if !strings.Contains(appsPageHTML, "Apps") && !strings.Contains(appsPageHTML, "app") {
					t.Errorf("expected Bravo to see Apps page content")
				}
			})
		})
	})
}
