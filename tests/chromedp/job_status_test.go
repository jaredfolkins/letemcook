package main_test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jaredfolkins/letemcook/tests/testutil"
)

// TestJobStatusPolling tests the job status endpoint that was failing in the HAR
func TestJobStatusPolling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load the actual squid values from the test environment
	alphaSquid, _, err := testutil.LoadTestEnv()
	if err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	// Create context with detailed logging
	ctx, cancel := createLoggedContext()
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)
	defer cancelTimeout()

	loginURLValues := url.Values{}
	loginURLValues.Set("squid", alphaSquid)
	loginURLValues.Set("account", testutil.AlphaAccountName)
	loginURL := testutil.GetBaseURL() + "/lemc/login" + "?" + loginURLValues.Encode()

	log.Printf("Starting job status test - Login URL: %s", loginURL)

	var finalURL string
	var pageHTML string
	var jobStatusHTML string

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
		chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second), // Wait for login to complete

		// Navigate to a cookbook edit page similar to the HAR
		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/cookbooks"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Location(&finalURL),
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
	}

	log.Println("Executing login and navigation tasks...")
	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed during login and navigation: %v", err)
	}

	log.Printf("After login - Final URL: %s", finalURL)
	log.Printf("Page contains 'Jobs': %v", strings.Contains(pageHTML, "Jobs"))
	log.Printf("Page contains job status elements: %v", strings.Contains(pageHTML, "job-status"))

	// Navigate to the jobs page to trigger job-related functionality
	jobsPageTasks := chromedp.Tasks{
		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/account/jobs"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
	}

	log.Println("Navigating to jobs page...")
	if err := chromedp.Run(ctx, jobsPageTasks); err != nil {
		t.Fatalf("failed navigating to jobs page: %v", err)
	}

	log.Printf("Jobs page HTML length: %d", len(pageHTML))
	log.Printf("Jobs page contains error: %v", strings.Contains(pageHTML, "error") || strings.Contains(pageHTML, "Error"))
	log.Printf("Jobs page contains 'No jobs': %v", strings.Contains(pageHTML, "No jobs"))

	// Try to find any cookbook with UUID to test job status endpoint
	var cookbookLink string
	findCookbookTasks := chromedp.Tasks{
		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/cookbooks"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Evaluate(`
			const links = Array.from(document.querySelectorAll('a[href*="/lemc/cookbook/edit/"]'));
			links.length > 0 ? links[0].href : '';
		`, &cookbookLink),
	}

	log.Println("Looking for cookbook links...")
	if err := chromedp.Run(ctx, findCookbookTasks); err != nil {
		t.Fatalf("failed finding cookbook links: %v", err)
	}

	log.Printf("Found cookbook link: %s", cookbookLink)

	if cookbookLink != "" {
		// Extract UUID from the link and test job status endpoint
		parts := strings.Split(cookbookLink, "/")
		if len(parts) > 0 {
			// Get the last part and remove any query parameters
			lastPart := parts[len(parts)-1]
			uuid := strings.Split(lastPart, "?")[0] // Remove query params
			log.Printf("Testing with UUID: %s", uuid)

			// Navigate to cookbook edit page
			editPageTasks := chromedp.Tasks{
				chromedp.Navigate(cookbookLink),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),
				chromedp.Sleep(2 * time.Second), // Wait for any auto-polling to start
			}

			log.Println("Navigating to cookbook edit page...")
			if err := chromedp.Run(ctx, editPageTasks); err != nil {
				t.Fatalf("failed navigating to cookbook edit page: %v", err)
			}

			// Test the job status endpoint directly (similar to HAR requests)
			jobStatusURL := fmt.Sprintf("%s/lemc/cookbook/job/status/uuid/%s/page/1/scope/individual",
				testutil.GetBaseURL(), uuid)

			log.Printf("Testing job status endpoint: %s", jobStatusURL)

			jobStatusTasks := chromedp.Tasks{
				chromedp.Navigate(jobStatusURL),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),
				chromedp.OuterHTML("html", &jobStatusHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, jobStatusTasks); err != nil {
				t.Fatalf("failed testing job status endpoint: %v", err)
			}

			log.Printf("Job status response length: %d", len(jobStatusHTML))
			log.Printf("Job status contains CPU icon: %v", strings.Contains(jobStatusHTML, "lemc-cpu-icon"))
			log.Printf("Job status contains clock icon: %v", strings.Contains(jobStatusHTML, "lemc-clock-icon"))
			log.Printf("Job status contains cron icon: %v", strings.Contains(jobStatusHTML, "lemc-cron-icon"))
			log.Printf("Job status HTML: %s", jobStatusHTML)

			// Check for the badge values
			if strings.Contains(jobStatusHTML, `indicator-item indicator-end indicator-middle badge">0<`) {
				log.Println("Found zero badge values - this matches the HAR issue")
			}

			// Check if this is an error response
			if strings.Contains(jobStatusHTML, "error") || strings.Contains(jobStatusHTML, "Error") {
				t.Errorf("Job status endpoint returned an error")
			}
		}
	} else {
		log.Println("No cookbook links found - creating a test cookbook")

		// Create a simple cookbook to test with
		createCookbookTasks := chromedp.Tasks{
			chromedp.Navigate(testutil.GetBaseURL() + "/lemc/cookbooks/create"),
			chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="name"]`, "Test Cookbook", chromedp.ByQuery),
			chromedp.SendKeys(`textarea[name="description"]`, "Test cookbook for job status", chromedp.ByQuery),
			chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
			chromedp.Sleep(2 * time.Second),
			chromedp.Location(&finalURL),
		}

		if err := chromedp.Run(ctx, createCookbookTasks); err != nil {
			t.Fatalf("failed creating test cookbook: %v", err)
		}

		log.Printf("Created cookbook, final URL: %s", finalURL)

		// Extract UUID from the final URL and test job status
		if strings.Contains(finalURL, "/lemc/cookbook/edit/") {
			parts := strings.Split(finalURL, "/")
			if len(parts) > 0 {
				uuid := parts[len(parts)-1]
				log.Printf("Testing job status with created cookbook UUID: %s", uuid)

				jobStatusURL := fmt.Sprintf("%s/lemc/cookbook/job/status/uuid/%s/page/1/scope/individual",
					testutil.GetBaseURL(), uuid)

				jobStatusTasks := chromedp.Tasks{
					chromedp.Navigate(jobStatusURL),
					chromedp.WaitVisible(`body`, chromedp.ByQuery),
					chromedp.OuterHTML("html", &jobStatusHTML, chromedp.ByQuery),
				}

				if err := chromedp.Run(ctx, jobStatusTasks); err != nil {
					t.Fatalf("failed testing job status endpoint with created cookbook: %v", err)
				}

				log.Printf("Job status response: %s", jobStatusHTML)
			}
		}
	}
}

// TestJobsPageFunctionality tests the main jobs page
func TestJobsPageFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	alphaSquid, _, err := testutil.LoadTestEnv()
	if err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	ctx, cancel := createLoggedContext()
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	loginURLValues := url.Values{}
	loginURLValues.Set("squid", alphaSquid)
	loginURLValues.Set("account", testutil.AlphaAccountName)
	loginURL := testutil.GetBaseURL() + "/lemc/login" + "?" + loginURLValues.Encode()

	var pageHTML string

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(testutil.UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(testutil.UsernameSelector, testutil.AlphaOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(testutil.PasswordSelector, testutil.TestPassword, chromedp.ByQuery),
		chromedp.Click(testutil.LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second),

		// Navigate directly to jobs page
		chromedp.Navigate(testutil.GetBaseURL() + "/lemc/account/jobs"),
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
}

// createLoggedContext creates a Chrome context with detailed logging
func createLoggedContext() (context.Context, context.CancelFunc) {
	// Configure Chrome options with additional logging
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("log-level", "0"),
		chromedp.Flag("v", "1"),
		chromedp.WindowSize(1920, 1080),
	)

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create Chrome context with logging
	ctx, ctxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Return a combined cancel function
	combinedCancel := func() {
		ctxCancel()
		allocCancel()
		time.Sleep(100 * time.Millisecond)
	}

	return ctx, combinedCancel
}
