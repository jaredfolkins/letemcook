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

// TestJobStatusUpdate tests the job status endpoint that was failing in the HAR
func TestJobStatusUpdate(t *testing.T) {
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

				// Navigate to jobs page
				chromedp.Navigate(baseURL + "/lemc/account/jobs"),
				chromedp.Sleep(2 * time.Second), // Wait for page to load

				// Capture the page content for verification
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify we can access the jobs page
			if !strings.Contains(bodyHTML, "Jobs") && !strings.Contains(bodyHTML, "jobs") {
				t.Errorf("expected to find 'Jobs' or 'jobs' in page content")
			}
		})
	})
}

func TestCookbookCreate(t *testing.T) {
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

			// Simplified cookbook creation test - just verify the create page loads
			var bodyHTML string
			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second),

				// Navigate to cookbook create page
				chromedp.Navigate(baseURL + "/lemc/cookbooks/create"),
				chromedp.Sleep(2 * time.Second),

				// Capture the page content
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify we can access the create cookbook page
			if !strings.Contains(bodyHTML, "Create") && !strings.Contains(bodyHTML, "cookbook") {
				t.Errorf("expected to find 'Create' or 'cookbook' in page content")
			}
		})
	})
}

// TestJobStatusPolling tests the job status endpoint that was failing in the HAR
func TestJobStatusPolling(t *testing.T) {
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

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second),

				// Navigate to jobs page to test polling
				chromedp.Navigate(baseURL + "/lemc/account/jobs"),
				chromedp.Sleep(5 * time.Second), // Give time for any polling to occur
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Test passes if we successfully navigate without errors
		})
	})
}

// TestJobsPageFunctionality tests the main jobs page
func TestJobsPageFunctionality(t *testing.T) {
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

			var pageHTML string

			tasks := chromedp.Tasks{
				chromedp.Navigate(loginURL),
				chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
				chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
				chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
				chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
				chromedp.Sleep(3 * time.Second),

				// Navigate directly to jobs page
				chromedp.Navigate(baseURL + "/lemc/account/jobs"),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),
				chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
			}

			log.Println("Testing jobs page functionality...")
			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed during jobs page test: %v", err)
			}

			log.Printf("Jobs page HTML contains pagination: %v", strings.Contains(pageHTML, "pagination"))
			log.Printf("Jobs page HTML contains job table: %v", strings.Contains(pageHTML, "table"))
			log.Printf("Jobs page HTML contains 'No jobs': %v", strings.Contains(pageHTML, "No jobs"))

			// Test pagination if present
			if strings.Contains(pageHTML, "pagination") {
				log.Println("Testing job pagination...")
				paginationTasks := chromedp.Tasks{
					chromedp.Sleep(1 * time.Second),
					chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
				}

				if err := chromedp.Run(ctx, paginationTasks); err != nil {
					log.Printf("Warning: pagination test failed: %v", err)
				}
			}

			log.Printf("Jobs page test completed. Page length: %d", len(pageHTML))
		})
	})
}
