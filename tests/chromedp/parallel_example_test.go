package main_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jaredfolkins/letemcook/tests/testutil"
)

func TestParallelCookbookCreateExample(t *testing.T) {
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

			t.Logf("Using instance %s on port %d", instance.ID, instance.Port)
			t.Logf("Login URL: %s", loginURL)

			var hasCreateButton bool

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second),

				// Navigate to cookbooks page
				chromedp.Navigate(baseURL + "/lemc/cookbooks"),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),

				// Check for create button
				chromedp.Evaluate(`document.querySelector('a[href*="create"]') !== null`, &hasCreateButton),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			if !hasCreateButton {
				t.Error("Expected to find create cookbook button")
			}
		})
	})
}

func TestParallelSimpleExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test can run in parallel with others
	t.Parallel()

	// Use simple test wrapper for non-ChromeDP tests
	testutil.SimpleTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Simple test that just verifies the server is responding
		baseURL := testutil.GetBaseURLForInstance(instance)

		t.Logf("Testing instance %s on port %d", instance.ID, instance.Port)
		t.Logf("Base URL: %s", baseURL)

		// Could add HTTP client tests here
		// client := &http.Client{Timeout: 5 * time.Second}
		// resp, err := client.Get(baseURL + "/")
		// ... test logic

		// For now, just verify we have a unique instance
		if instance.Port < 15362 {
			t.Errorf("Expected port >= 15362, got %d", instance.Port)
		}

		if len(instance.ID) != 8 {
			t.Errorf("Expected 8-character ID, got %s", instance.ID)
		}
	})
}
