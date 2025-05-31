package tests

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestChromeDPLoginFailure(t *testing.T) {
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

			log.Printf("Redirected to login URL: %s", finalURL)
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

			log.Println("Login failure test completed as expected (login form likely still present).")
		})
	})
}

func TestChromeDPSuccessfulLogins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		// Each login test gets its own ChromeDP context for complete isolation
		t.Run("AlphaOwnerLogin", func(t *testing.T) {
			ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				err := LoginToAlphaAccount(ctx, instance)
				if err != nil {
					t.Fatalf("failed running Alpha login: %v", err)
				}

				var title string
				err = chromedp.Run(ctx, chromedp.Title(&title))
				if err != nil {
					t.Fatalf("failed to get page title: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, AlphaAccountName) {
					t.Errorf("expected title to contain '%s', got: %s", AlphaAccountName, title)
				}
			})
		})

		t.Run("BravoOwnerLogin", func(t *testing.T) {
			ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
				err := LoginToBravoAccount(ctx, instance)
				if err != nil {
					t.Fatalf("failed running Bravo login: %v", err)
				}

				var title string
				err = chromedp.Run(ctx, chromedp.Title(&title))
				if err != nil {
					t.Fatalf("failed to get page title: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, BravoAccountName) {
					t.Errorf("expected title to contain '%s', got: %s", BravoAccountName, title)
				}
			})
		})
	})
}
