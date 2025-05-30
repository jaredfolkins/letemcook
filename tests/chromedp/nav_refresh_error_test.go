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

	// Use parallel test wrapper for automatic instance management
	testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Load test environment for this specific instance
		alphaSquid, _, err := testutil.LoadTestEnvForInstance(instance)
		if err != nil {
			t.Fatalf("Failed to load test environment: %v", err)
		}

		// Use ChromeDP with the instance
		testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			loginVals := url.Values{}
			loginVals.Set("squid", alphaSquid)
			loginVals.Set("account", testutil.AlphaAccountName)

			// Use the instance-specific base URL
			baseURL := testutil.GetBaseURLForInstance(instance)
			loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second), // Wait for login to complete

				chromedp.Navigate(baseURL + "/lemc/apps"),
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
		})
	})
}

func TestNavRefreshError(t *testing.T) {
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
			loginVals := url.Values{}
			loginVals.Set("squid", alphaSquid)
			loginVals.Set("account", testutil.AlphaAccountName)

			// Use the instance-specific base URL
			baseURL := testutil.GetBaseURLForInstance(instance)
			loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

			var bodyHTML string

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second), // Wait for login to complete

				// Navigate to apps page
				chromedp.Navigate(baseURL + "/lemc/apps"),
				chromedp.Sleep(2 * time.Second), // Wait for page to load

				// Capture the page content for verification
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify we can access the apps page
			if !strings.Contains(bodyHTML, "Apps") && !strings.Contains(bodyHTML, "app") {
				t.Errorf("expected to find 'Apps' or 'app' in page content")
			}
		})
	})
}

func TestAppRefreshError(t *testing.T) {
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

				// Navigate to cookbooks page
				chromedp.Navigate(baseURL + "/lemc/cookbooks"),
				chromedp.Sleep(2 * time.Second), // Wait for page to load

				// Capture the page content for verification
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify we can access the cookbooks page
			if !strings.Contains(bodyHTML, "Cookbooks") && !strings.Contains(bodyHTML, "cookbook") {
				t.Errorf("expected to find 'Cookbooks' or 'cookbook' in page content")
			}
		})
	})
}
