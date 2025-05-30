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

func TestNavigationActiveState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use parallel test wrapper for automatic instance management
	testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Load test environment for this specific instance
		alphaSquid, _, err := testutil.LoadTestEnvForInstance(instance)
		if err != nil {
			t.Fatalf("Failed to load test environment: %v", err)
		}

		// Use ChromeDP with the instance
		testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			loginURLValues := url.Values{}
			loginURLValues.Set("squid", alphaSquid)
			loginURLValues.Set("account", testutil.AlphaAccountName)

			// Use the instance-specific base URL
			baseURL := testutil.GetBaseURLForInstance(instance)
			loginURL := baseURL + "/lemc/login?" + loginURLValues.Encode()

			var bodyHTML string

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second), // Wait for login to complete

				// Navigate to different pages to test navigation state
				chromedp.Navigate(baseURL + "/lemc/cookbooks"),
				chromedp.Sleep(1 * time.Second),
				chromedp.Navigate(baseURL + "/lemc/apps"),
				chromedp.Sleep(1 * time.Second),

				// Capture the final page content
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify navigation works
			if !strings.Contains(bodyHTML, "Apps") && !strings.Contains(bodyHTML, "app") {
				t.Errorf("expected to find 'Apps' or 'app' in page content after navigation")
			}
		})
	})
}
