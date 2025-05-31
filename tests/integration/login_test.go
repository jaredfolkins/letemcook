package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestActualLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			baseURL := GetBaseURLForInstance(instance)
			startURL := baseURL + "/"

			var finalURL string

			tasks := chromedp.Tasks{
				chromedp.Navigate(startURL),
				chromedp.WaitVisible(UsernameSelector, chromedp.ByQuery),
				chromedp.Location(&finalURL),
				chromedp.SendKeys(UsernameSelector, "invaliduser", chromedp.ByQuery),
				chromedp.SendKeys(PasswordSelector, "wrongpassword", chromedp.ByQuery),
				chromedp.Click(LoginButtonSelector, chromedp.ByQuery),
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

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		// Test Alpha account login
		t.Run("Alpha", func(t *testing.T) {
			ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				err := LoginToAlphaAccount(ctx, instance)
				if err != nil {
					t.Fatalf("failed to login to Alpha account: %v", err)
				}

				var title string
				err = chromedp.Run(ctx, chromedp.Title(&title))
				if err != nil {
					t.Fatalf("failed to get page title: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Alpha") {
					t.Errorf("expected title to contain 'Account Alpha', got: %s", title)
				}
			})
		})

		// Test Bravo account login - gets its own fresh browser context
		t.Run("Bravo", func(t *testing.T) {
			ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				err := LoginToBravoAccount(ctx, instance)
				if err != nil {
					t.Fatalf("failed to login to Bravo account: %v", err)
				}

				var title string
				err = chromedp.Run(ctx, chromedp.Title(&title))
				if err != nil {
					t.Fatalf("failed to get page title: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Bravo") {
					t.Errorf("expected title to contain 'Account Bravo', got: %s", title)
				}
			})
		})
	})
}
