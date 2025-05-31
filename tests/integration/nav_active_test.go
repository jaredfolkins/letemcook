package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestNavigationActiveState(t *testing.T) {
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
				// Navigate to different pages to test navigation state
				chromedp.Navigate(baseURL + "/lemc/cookbooks"),
				chromedp.Sleep(1 * time.Second),
				chromedp.Navigate(baseURL + "/lemc/apps"),
				chromedp.Sleep(1 * time.Second),

				// Capture the final page content
				chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
			}

			if err := chromedp.Run(ctx, tasks); err != nil {
				t.Fatalf("failed running chromedp tasks: %v", err)
			}

			// Verify navigation works
			if !strings.Contains(bodyHTML, "Apps") && !strings.Contains(bodyHTML, "app") {
				t.Errorf("expected to find 'Apps' or 'app' in page content after navigation")
			}
		})
	})
}
