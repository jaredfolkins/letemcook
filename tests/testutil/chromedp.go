package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/sqids/sqids-go"
)

// Test credentials that match the seed data
const (
	AlphaOwnerUsername = "alpha-owner"  // matches seed data
	BravoOwnerUsername = "bravo-owner"  // matches seed data
	TestPassword       = "asdfasdfasdf" // matches the bcrypt hash in seed data
	AlphaAccountName   = "Account Alpha"
	BravoAccountName   = "Account Bravo"
	AlphaAccountID     = 1 // from seed data
	BravoAccountID     = 2 // from seed data
)

// Common selectors
const (
	UsernameSelector     = `input[name="username"]`
	PasswordSelector     = `input[name="password"]`
	LoginButtonSelector  = `button.btn-primary`
	FlashSuccessSelector = `.toast-alerts .alert-success`
	FlashErrorSelector   = `.toast-alerts .alert-error`
)

// LoadTestEnv loads the test environment file and returns the squid values
func LoadTestEnv() (alphaSquid, bravoSquid string, err error) {
	// Load the test .env file
	envPath := filepath.Join(DataRoot(), "test", ".env")
	if err := godotenv.Load(envPath); err != nil {
		return "", "", err
	}

	// Get the alphabet from the environment
	alphabet := os.Getenv("LEMC_SQUID_ALPHABET")
	if alphabet == "" {
		return "", "", fmt.Errorf("LEMC_SQUID_ALPHABET not found in environment")
	}

	// Create squid generator
	s, err := sqids.New(sqids.Options{
		Blocklist: nil,
		MinLength: 4,
		Alphabet:  alphabet,
	})
	if err != nil {
		return "", "", err
	}

	// Generate squids for account IDs 1 and 2
	alphaSquid, err = s.Encode([]uint64{AlphaAccountID})
	if err != nil {
		return "", "", err
	}

	bravoSquid, err = s.Encode([]uint64{BravoAccountID})
	if err != nil {
		return "", "", err
	}

	return alphaSquid, bravoSquid, nil
}

// GetBaseURL returns the test server base URL
func GetBaseURL() string {
	port := os.Getenv("LEMC_PORT_TEST")
	if port == "" {
		port = "15362"
	}
	return "http://localhost:" + port
}

// CreateHeadlessContext creates a headless Chrome context for testing with proper cleanup
func CreateHeadlessContext() (context.Context, context.CancelFunc) {
	// Configure Chrome options for testing
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.WindowSize(1920, 1080),
	)

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create Chrome context
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// Return a combined cancel function that cleans up both contexts
	combinedCancel := func() {
		ctxCancel()
		allocCancel()

		// Give Chrome a moment to shutdown gracefully
		time.Sleep(100 * time.Millisecond)
	}

	return ctx, combinedCancel
}
