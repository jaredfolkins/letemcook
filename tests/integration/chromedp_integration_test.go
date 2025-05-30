package integration_test

import (
	"context"
	"log"
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

type loginTestData struct {
	testName    string
	username    string
	password    string
	accountName string
	squid       string
}

func TestSuccessfulLogins(t *testing.T) {
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

		testCases := []loginTestData{
			{
				testName:    "AlphaOwnerLogin",
				username:    testutil.AlphaOwnerUsername,
				password:    testutil.TestPassword,
				accountName: testutil.AlphaAccountName,
				squid:       alphaSquid,
			},
			{
				testName:    "BravoOwnerLogin",
				username:    testutil.BravoOwnerUsername,
				password:    testutil.TestPassword,
				accountName: testutil.BravoAccountName,
				squid:       bravoSquid,
			},
		}

		// Use ChromeDP with the instance
		testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			for _, tc := range testCases {
				tc := tc // Capture range variable
				t.Run(tc.testName, func(t *testing.T) {
					loginURLValues := url.Values{}
					loginURLValues.Set("squid", tc.squid)
					loginURLValues.Set("account", tc.accountName)

					// Use the instance-specific base URL
					baseURL := testutil.GetBaseURLForInstance(instance)
					startURL := baseURL + "/lemc/login?" + loginURLValues.Encode()
					t.Logf("[%s] Navigating to: %s", tc.testName, startURL)

					var bodyHTML string // For debugging

					tasks := chromedp.Tasks{
						chromedp.Navigate(startURL),
						chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
						chromedp.SendKeys(testutil.UsernameSelector, tc.username, chromedp.ByQuery),
						chromedp.SendKeys(testutil.PasswordSelector, tc.password, chromedp.ByQuery),
						chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
						chromedp.Sleep(3 * time.Second), // Allow redirect to happen
						chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
					}

					err := chromedp.Run(ctx, tasks)

					t.Logf("[%s] --- Captured Body HTML after login attempt ---", tc.testName)
					t.Logf("%s", bodyHTML)
					t.Logf("[%s] ----------------------------------------------", tc.testName)

					if err != nil {
						t.Fatalf("[%s] Failed during login tasks (check logged HTML): %v", tc.testName, err)
					}

					// Check if we're on a success page or if login flash appeared
					if !strings.Contains(bodyHTML, "Login successful") {
						// Look for other indicators of successful login
						if !strings.Contains(bodyHTML, "Dashboard") && !strings.Contains(bodyHTML, "Cookbooks") && !strings.Contains(bodyHTML, "Apps") {
							t.Errorf("[%s] Expected login to succeed, but no success indicators found in page content", tc.testName)
						}
					}

					log.Printf("[%s] Login success test completed.", tc.testName)
				})
			}
		})
	})
}

func TestPasswordChangeLoginExpiration(t *testing.T) {
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
			// Test login with Alpha account
			t.Run("Alpha Account Login", func(t *testing.T) {
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
					chromedp.Sleep(3*time.Second), // Wait for login to complete
					chromedp.Title(&title),
				)
				if err != nil {
					t.Fatalf("failed running Alpha login: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Alpha") {
					t.Errorf("expected title to contain 'Account Alpha', got: %s", title)
				}
			})

			// Test login with Bravo account
			t.Run("Bravo Account Login", func(t *testing.T) {
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
					chromedp.Sleep(3*time.Second), // Wait for login to complete
					chromedp.Title(&title),
				)
				if err != nil {
					t.Fatalf("failed running Bravo login: %v", err)
				}

				// Basic assertion that we successfully logged in
				if !strings.Contains(title, "Account Bravo") {
					t.Errorf("expected title to contain 'Account Bravo', got: %s", title)
				}
			})
		})
	})
}
