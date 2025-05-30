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

func TestActualLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use parallel test wrapper for automatic instance management
	testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Use ChromeDP with the instance
		testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			// Use the instance-specific base URL
			baseURL := testutil.GetBaseURLForInstance(instance)
			startURL := baseURL + "/"

			var finalURL string

			tasks := chromedp.Tasks{
				chromedp.Navigate(startURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.Location(&finalURL),
				chromedp.SendKeys(testutil.UsernameSelector, "invaliduser", chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, "wrongpassword", chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(1 * time.Second),
			}

			err := chromedp.Run(ctx, tasks)
			if err != nil {
				t.Fatalf("Failed to run login failure tasks: %v", err)
			}

			t.Logf("Redirected to login URL: %s", finalURL)
			if !strings.Contains(finalURL, "/lemc/login") {
				t.Errorf("Expected navigation to root to redirect to a URL containing '/lemc/login', but got: %s", finalURL)
			}
			if !strings.Contains(finalURL, "squid=") || !strings.Contains(finalURL, "account=") {
				t.Errorf("Expected redirected URL '%s' to contain 'squid' and 'account' query parameters", finalURL)
			}

			// Simple check for password field visibility
			var passwordInputVisible bool
			err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector('input[name="password"]') !== null`, &passwordInputVisible))

			if err != nil {
				t.Fatalf("Failed to query password input visibility after failed login attempt: %v", err)
			}

			if !passwordInputVisible {
				t.Errorf("Expected login to fail (password input should still be visible), but it seemed to succeed or the element disappeared unexpectedly.")
			}

			t.Logf("Login failure test completed as expected (login form likely still present).")
		})
	})
}

type loginTestData struct {
	testName    string
	username    string
	password    string
	accountName string
	squid       string
}

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use parallel test wrapper for automatic instance management
	testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		// Load test environment using instance environment variables
		alphaSquid, bravoSquid, err := testutil.LoadTestEnvForInstance(instance)
		if err != nil {
			t.Fatalf("failed loading test environment: %v", err)
		}

		// Test Alpha account login
		t.Run("Alpha", func(t *testing.T) {
			testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				loginVals := url.Values{}
				loginVals.Set("squid", alphaSquid)
				loginVals.Set("account", testutil.AlphaAccountName)

				// Use the instance-specific base URL
				baseURL := testutil.GetBaseURLForInstance(instance)
				loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

				var title string
				err := chromedp.Run(ctx,
					chromedp.Navigate(loginURL),
					chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
					chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
					chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
					chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
					chromedp.Sleep(3*time.Second), // Wait for redirect/login to complete
					chromedp.Title(&title),
				)
				if err != nil {
					t.Fatalf("failed running chromedp tasks: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Alpha") {
					t.Errorf("expected title to contain 'Account Alpha', got: %s", title)
				}
			})
		})

		// Test Bravo account login - gets its own fresh browser context
		t.Run("Bravo", func(t *testing.T) {
			testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				loginVals := url.Values{}
				loginVals.Set("squid", bravoSquid)
				loginVals.Set("account", testutil.BravoAccountName)

				// Use the instance-specific base URL
				baseURL := testutil.GetBaseURLForInstance(instance)
				loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

				var title string
				err := chromedp.Run(ctx,
					chromedp.Navigate(loginURL),
					chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
					chromedp.SendKeys(testutil.UsernameSelector, testutil.BravoOwnerUsername, chromedp.ByQuery),
					chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
					chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
					chromedp.Sleep(3*time.Second), // Wait for redirect/login to complete
					chromedp.Title(&title),
				)
				if err != nil {
					t.Fatalf("failed running chromedp tasks: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Bravo") {
					t.Errorf("expected title to contain 'Account Bravo', got: %s", title)
				}
			})
		})
	})
}
