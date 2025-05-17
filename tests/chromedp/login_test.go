package main_test

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func getBaseURL() string {
	port := os.Getenv("LEMC_PORT")
	if port == "" {
		port = "5362"
	}
	return "http://localhost:" + port
}

var (
	baseURL              = getBaseURL()
	loginPath            = "/lemc/login"
	validUsername        = "alpha_owner"
	validPassword        = "asdfasdfasdf"
	flashSuccessSelector = `.toast-alerts .alert-success` // Corrected selector
	flashErrorSelector   = `.toast-alerts .alert-error`   // Corrected selector
)

const (
	usernameSelector    = `input[name="username"]`
	passwordSelector    = `input[name="password"]`
	loginButtonSelector = `button.btn-primary`
)

func createHeadlessContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.Flag("disable-gpu", true),
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancelCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	cancelAll := func() {
		cancelCtx()
		cancelAlloc()
	}

	return ctx, cancelAll
}

func TestActualLoginFailure(t *testing.T) {
	ctx, cancel := createHeadlessContext(t)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 20*time.Second)
	defer cancelTimeout()

	startURL := baseURL + "/"

	var finalURL string

	tasks := chromedp.Tasks{
		chromedp.Navigate(startURL),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.Location(&finalURL),
		chromedp.SendKeys(usernameSelector, "invaliduser", chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, "wrongpassword", chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(500 * time.Millisecond),
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

	var passwordInputVisible bool
	err = chromedp.Run(ctx, chromedp.QueryAfter(passwordSelector, func(ctx context.Context, execID runtime.ExecutionContextID, nodes ...*cdp.Node) error {
		_ = execID
		passwordInputVisible = len(nodes) > 0
		return nil
	}, chromedp.ByQuery, chromedp.AtLeast(0)))

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
	testCases := []loginTestData{
		{
			testName:    "AlphaOwnerLogin",
			username:    validUsername, // "alpha_owner"
			password:    validPassword, // "asdfasdfasdf"
			accountName: "Account Alpha",
			squid:       "52wJ",
		},
		{
			testName:    "BravoOwnerLogin",
			username:    "bravo_owner",
			password:    validPassword, // Assuming same password "asdfasdfasdf"
			accountName: "Account Bravo",
			squid:       "vpOd",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			ctx, cancel := createHeadlessContext(t)
			defer cancel()

			ctx, cancelTimeout := context.WithTimeout(ctx, 20*time.Second)
			defer cancelTimeout()

			loginURLValues := url.Values{}
			loginURLValues.Set("squid", tc.squid)
			loginURLValues.Set("account", tc.accountName)
			startURL := baseURL + loginPath + "?" + loginURLValues.Encode()
			t.Logf("[%s] Navigating to: %s", tc.testName, startURL)

			var successFlashText string
			var errorFlashText string // Keep checking for errors just in case
			var bodyHTML string       // For debugging

			tasks := chromedp.Tasks{
				chromedp.Navigate(startURL),
				chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(usernameSelector, tc.username, chromedp.ByQuery),
				chromedp.SendKeys(passwordSelector, tc.password, chromedp.ByQuery),
				chromedp.Click(loginButtonSelector, chromedp.ByQuery),

				chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),

				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),

				chromedp.Text(flashSuccessSelector, &successFlashText, chromedp.ByQuery),
			}

			err := chromedp.Run(ctx, tasks)

			t.Logf("[%s] --- Captured Body HTML after login attempt ---", tc.testName)
			t.Logf("%s", bodyHTML)
			t.Logf("[%s] ----------------------------------------------", tc.testName)

			t.Logf("[%s] --- Captured DOM Flash Messages ---", tc.testName)
			t.Logf("[%s] Success Flash Text (DOM): %q", tc.testName, successFlashText)
			t.Logf("[%s] Error Flash Text   (DOM): %q", tc.testName, errorFlashText)
			t.Logf("[%s] ---------------------------------", tc.testName)

			if err != nil {
				t.Fatalf("[%s] Failed during login tasks or waiting for success indication (check logged HTML and flashes): %v", tc.testName, err)
			}

			expectedFlashText := "Login successful."
			if !strings.Contains(successFlashText, expectedFlashText) {
				t.Errorf("[%s] Expected SUCCESS flash message text (from DOM) to contain %q, but got %q (Error flash was: %q)", tc.testName, expectedFlashText, successFlashText, errorFlashText)
			}

			log.Printf("[%s] Login success test completed.", tc.testName)
		})
	}
}
