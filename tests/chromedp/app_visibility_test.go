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

	// Load the actual squid values from the test environment
	alphaSquid, bravoSquid, err := testutil.LoadTestEnv()
	if err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	testCases := []appVisibilityTestData{
		{
			testName:           "AlphaOwnerSeesAlphaApp",
			username:           accAlphaOwnerUser,
			password:           seedPassword,
			accountName:        accAlphaName,
			squid:              alphaSquid,
			shouldSeeSubstr:    "Description for Alpha App", // Check description
			shouldNotSeeSubstr: bravoAppPrefix,
			expectPresence:     true,
		},
		{
			testName:           "AlphaViewerShouldSeeApps", // Use actual user from seed data
			username:           "alpha-viewer",             // Actual user in seed data
			password:           seedPassword,
			accountName:        accAlphaName,
			squid:              alphaSquid,
			shouldSeeSubstr:    alphaAppPrefix, // Regular user should see apps they have permission for
			shouldNotSeeSubstr: bravoAppPrefix, // Should not see other account's apps
			expectPresence:     true,           // This user has view permissions
		},
		{
			testName:           "BravoOwnerSeesBravoApp",
			username:           accBravoOwnerUser,
			password:           seedPassword,
			accountName:        accBravoName,
			squid:              bravoSquid,
			shouldSeeSubstr:    "Description for Bravo App", // Check description
			shouldNotSeeSubstr: alphaAppPrefix,
			expectPresence:     true,
		},
		{
			testName:           "BravoMainShouldSeeApps", // Use actual user from seed data
			username:           "bravo-main",             // Actual user in seed data
			password:           seedPassword,
			accountName:        accBravoName,
			squid:              bravoSquid,
			shouldSeeSubstr:    bravoAppPrefix, // Regular user should see apps they have permission for
			shouldNotSeeSubstr: alphaAppPrefix, // Should not see other account's apps
			expectPresence:     true,           // This user has view permissions
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.testName, func(t *testing.T) {

			ctx, cancel := testutil.CreateHeadlessContext()
			defer cancel()

			ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second) // Increased timeout
			defer cancelTimeout()

			loginURLValues := url.Values{}
			loginURLValues.Set("squid", tc.squid)
			loginURLValues.Set("account", tc.accountName)
			loginURL := testutil.GetBaseURL() + "/lemc/login" + "?" + loginURLValues.Encode()
			targetappsURL := testutil.GetBaseURL() + appsPath // URL to navigate to after login

			t.Logf("[%s] Navigating to login: %s", tc.testName, loginURL)

			var appListHTML string // Capture HTML of the app list for debugging

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, tc.username, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, tc.password, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(1 * time.Second),
				chromedp.Navigate(targetappsURL),
				chromedp.ActionFunc(func(ctx context.Context) error {
					captureSelector := `div.bg-base-100.p-9.edges.gap-12.mx-12.my-4` // Selector for the main content area
					if tc.expectPresence {
						return chromedp.Tasks{
							chromedp.WaitVisible(`tbody tr`, chromedp.ByQuery),
							chromedp.InnerHTML(captureSelector, &appListHTML, chromedp.ByQuery),
						}.Do(ctx)
					} else {
						return chromedp.Tasks{
							chromedp.Sleep(1 * time.Second),
							chromedp.InnerHTML(captureSelector, &appListHTML, chromedp.ByQuery),
						}.Do(ctx)
					}
				}),
			}

			err := chromedp.Run(ctx, tasks)
			if err != nil {
				t.Fatalf("[%s] Failed during browser automation tasks: %v", tc.testName, err)
			}

			t.Logf("[%s] --- Captured app List HTML ---", tc.testName)
			t.Logf("%s", appListHTML)
			t.Logf("[%s] ---------------------------------", tc.testName)

			log.Printf("[%s] Verifying app visibility for user %s...", tc.testName, tc.username)

			if tc.expectPresence {
				if !strings.Contains(appListHTML, tc.shouldSeeSubstr) {
					t.Errorf("[%s] Expected to find apps containing %q for user %s, but they were not found in the list.",
						tc.testName, tc.shouldSeeSubstr, tc.username)
					t.Logf("[%s] Captured List HTML:\\n%s", tc.testName, appListHTML) // Log HTML on failure
				} else {
					log.Printf("[%s]   ✅ Found expected substring: %q", tc.testName, tc.shouldSeeSubstr)
				}
			} else {
				if strings.Contains(appListHTML, tc.shouldSeeSubstr) {
					t.Errorf("[%s] Expected *not* to find apps containing %q for user %s (due to current behavior), but they were present.",
						tc.testName, tc.shouldSeeSubstr, tc.username)
					t.Logf("[%s] Captured List HTML:\\n%s", tc.testName, appListHTML) // Log HTML on failure
				} else {
					log.Printf("[%s]   ✅ Did not find substring %q as expected (current behavior).", tc.testName, tc.shouldSeeSubstr)
				}
			}

			if strings.Contains(appListHTML, tc.shouldNotSeeSubstr) {
				t.Errorf("[%s] Expected *not* to find apps containing %q for user %s, but they were present in the list.",
					tc.testName, tc.shouldNotSeeSubstr, tc.username)
				t.Logf("[%s] Captured List HTML:\\n%s", tc.testName, appListHTML) // Log HTML on failure
			} else {
				log.Printf("[%s]   ✅ Did not find unexpected substring: %q", tc.testName, tc.shouldNotSeeSubstr)
			}

			log.Printf("[%s] app visibility test completed.", tc.testName)
		})
	}
}
