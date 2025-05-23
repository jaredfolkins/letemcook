package main_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestNavActiveSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := createHeadlessContext(t)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 20*time.Second)
	defer cancelTimeout()

	loginURLValues := url.Values{}
	loginURLValues.Set("squid", "xkQN")
	loginURLValues.Set("account", "Account Alpha")
	loginURL := baseURL + loginPath + "?" + loginURLValues.Encode()

	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(usernameSelector, validUsername, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, validPassword, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitVisible(flashSuccessSelector, chromedp.ByQuery),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Fatalf("failed login flow: %v", err)
	}

	var appsClass string
	if err := chromedp.Run(ctx,
		chromedp.AttributeValue(`#navtop a[href="/lemc/apps?partial=true"]`, "class", &appsClass, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("read apps class: %v", err)
	}
	if !strings.Contains(appsClass, "active") {
		t.Fatalf("expected Apps nav to be active after login, got %q", appsClass)
	}

	if err := chromedp.Run(ctx,
		chromedp.Click(`#navtop a[href="/lemc/cookbooks?partial=true"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	); err != nil {
		t.Fatalf("click cookbooks: %v", err)
	}

	var cbClass string
	if err := chromedp.Run(ctx,
		chromedp.AttributeValue(`#navtop a[href="/lemc/cookbooks?partial=true"]`, "class", &cbClass, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("read cookbooks class: %v", err)
	}
	if !strings.Contains(cbClass, "active") {
		t.Fatalf("expected Cookbooks nav to be active after click, got %q", cbClass)
	}

	if err := chromedp.Run(ctx,
		chromedp.Click(`#navtop a[href="/lemc/apps?partial=true"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	); err != nil {
		t.Fatalf("click apps: %v", err)
	}

	appsClass = ""
	if err := chromedp.Run(ctx,
		chromedp.AttributeValue(`#navtop a[href="/lemc/apps?partial=true"]`, "class", &appsClass, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("read apps class after click: %v", err)
	}
	if !strings.Contains(appsClass, "active") {
		t.Fatalf("expected Apps nav to be active after click, got %q", appsClass)
	}
}
