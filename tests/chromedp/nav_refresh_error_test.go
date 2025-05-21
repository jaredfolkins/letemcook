package main_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// TestNavClickAfterHardRefresh captures JavaScript errors that occur when
// a user performs a hard refresh and then clicks a navigation link.
func TestNavClickAfterHardRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := createHeadlessContext(t)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var jsErr string
	chromedp.ListenTarget(ctx, func(ev any) {
		if exc, ok := ev.(*runtime.EventExceptionThrown); ok {
			if exc.ExceptionDetails != nil {
				desc := ""
				if exc.ExceptionDetails.Exception != nil {
					desc = exc.ExceptionDetails.Exception.Description
				}
				jsErr = exc.ExceptionDetails.Text + ": " + desc
			}
		}
	})

	loginVals := url.Values{}
	loginVals.Set("squid", "xkQN")
	loginVals.Set("account", "Account Alpha")
	loginURL := baseURL + loginPath + "?" + loginVals.Encode()

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(usernameSelector, validUsername, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, validPassword, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),

		chromedp.Navigate(baseURL + "/lemc/apps"),
		chromedp.WaitVisible(`#navtop`, chromedp.ByQuery),

		chromedp.ActionFunc(func(ctx context.Context) error {
			// Hard refresh by reloading the page while ignoring cache.
			return page.Reload().WithIgnoreCache(true).Do(ctx)
		}),
		chromedp.WaitVisible(`#navtop`, chromedp.ByQuery),
		chromedp.Click(`#navtop a[href="/lemc/cookbooks?partial=true"]`, chromedp.ByQuery),
		chromedp.Sleep(1 * time.Second),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed running chromedp tasks: %v", err)
	}

	if jsErr != "" {
		t.Fatalf("unexpected JavaScript error captured: %s", jsErr)
	}
}
