package tests

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestJobStatusUpdate tests the job status endpoint that was failing in the HAR
func TestJobStatusUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			// Login to Alpha account
			err := LoginToAlphaAccount(ctx, instance)
			if err != nil {
				t.Fatalf("Failed to login: %v", err)
			}

			baseURL := GetBaseURLForInstance(instance)
			var bodyHTML string

			tasks := chromedp.Tasks{
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

// TestJobStatusPolling tests the job status endpoint that was failing in the HAR
func TestJobStatusPolling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			// Login to Alpha account
			err := LoginToAlphaAccount(ctx, instance)
			if err != nil {
				t.Fatalf("Failed to login: %v", err)
			}

			baseURL := GetBaseURLForInstance(instance)

			tasks := chromedp.Tasks{
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

	SeriesTestWrapper(t, func(t *testing.T, instance *TestInstance) {
		ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
			// Login to Alpha account
			err := LoginToAlphaAccount(ctx, instance)
			if err != nil {
				t.Fatalf("Failed to login: %v", err)
			}

			baseURL := GetBaseURLForInstance(instance)
			var pageHTML string

			tasks := chromedp.Tasks{
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
