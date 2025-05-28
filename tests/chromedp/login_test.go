package main_test

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

var (
	baseURL   = testutil.GetBaseURL()
	loginPath = "/lemc/login"
)

func TestActualLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := testutil.CreateHeadlessContext()
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

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
	if !strings.Contains(finalURL, loginPath) {
		t.Errorf("Expected navigation to root to redirect to a URL containing '%s', but got: %s", loginPath, finalURL)
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

	// Load the actual squid values from the test environment
	alphaSquid, bravoSquid, err := testutil.LoadTestEnv()
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

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.testName, func(t *testing.T) {

			ctx, cancel := testutil.CreateHeadlessContext()
			defer cancel()

			ctx, cancelTimeout := context.WithTimeout(ctx, 20*time.Second)
			defer cancelTimeout()

			loginURLValues := url.Values{}
			loginURLValues.Set("squid", tc.squid)
			loginURLValues.Set("account", tc.accountName)
			startURL := baseURL + loginPath + "?" + loginURLValues.Encode()
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
}
