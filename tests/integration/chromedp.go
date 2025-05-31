package tests

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"testing"

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

var (
	chromeDPCleanupMutex sync.Mutex
	activeContexts       []*chromeDPContext
	testPackageComplete  = make(chan struct{})
	cleanupOnce          sync.Once
)

type chromeDPContext struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

// init sets up package-level cleanup
func init() {
	// Start a goroutine that will force cleanup when tests are done
	go func() {
		<-testPackageComplete
		time.Sleep(500 * time.Millisecond) // Give normal cleanup a chance

		// Force cleanup everything
		ForceCleanupChrome()

		// Force garbage collection
		runtime.GC()
		runtime.GC()

		// Give a final moment for cleanup
		time.Sleep(200 * time.Millisecond)
	}()
}

// MarkTestPackageComplete should be called when all tests are done
func MarkTestPackageComplete() {
	cleanupOnce.Do(func() {
		close(testPackageComplete)
	})
}

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

// createHeadlessContext creates a headless Chrome context for testing with proper cleanup
func createHeadlessContext() (context.Context, context.CancelFunc) {
	chromeDPCleanupMutex.Lock()
	defer chromeDPCleanupMutex.Unlock()

	// Configure Chrome options for testing with aggressive cleanup
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("force-device-scale-factor", "1"),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-component-update", true),
		chromedp.Flag("disable-domain-reliability", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.WindowSize(1920, 1080),
	)

	// Create allocator context with timeout
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create Chrome context with timeout
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// Track this context for global cleanup
	cdpCtx := &chromeDPContext{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		ctx:         ctx,
		cancel:      ctxCancel,
	}
	activeContexts = append(activeContexts, cdpCtx)

	// Enhanced cancel function that ensures proper cleanup
	combinedCancel := func() {
		chromeDPCleanupMutex.Lock()
		defer chromeDPCleanupMutex.Unlock()

		// Remove from active contexts
		for i, active := range activeContexts {
			if active == cdpCtx {
				activeContexts = append(activeContexts[:i], activeContexts[i+1:]...)
				break
			}
		}

		// Cancel the Chrome context first
		ctxCancel()

		// Create a timeout context for cleanup operations
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()

		// Try to gracefully shut down Chrome
		done := make(chan struct{})
		go func() {
			defer close(done)
			allocCancel()
		}()

		// Wait for graceful shutdown or timeout
		select {
		case <-done:
			// Graceful shutdown completed
		case <-cleanupCtx.Done():
			// Timeout - force cleanup
			allocCancel()
		}

		// Give extra time for all processes to terminate
		time.Sleep(100 * time.Millisecond)
	}

	return ctx, combinedCancel
}

// ForceCleanupChrome attempts to kill any hanging Chrome processes (emergency cleanup)
func ForceCleanupChrome() {
	chromeDPCleanupMutex.Lock()
	defer chromeDPCleanupMutex.Unlock()

	// Cancel all active contexts
	for _, cdpCtx := range activeContexts {
		cdpCtx.cancel()
		cdpCtx.allocCancel()
	}
	activeContexts = nil

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Force garbage collection to help with cleanup
	runtime.GC()
	runtime.GC() // Call twice to be thorough
}

// ChromeDPTestWrapperWithInstance provides a wrapper for ChromeDP tests with a specific test instance
func ChromeDPTestWrapperWithInstance(t *testing.T, instance *TestInstance, testFunc func(context.Context)) {
	ctx, cancel := createHeadlessContext()

	// Ensure cleanup happens even if test panics
	defer func() {
		if r := recover(); r != nil {
			cancel()
			time.Sleep(100 * time.Millisecond)
			panic(r)
		}
		cancel()
		// Extra cleanup time to ensure Chrome shuts down
		time.Sleep(150 * time.Millisecond)
	}()

	// Add test timeout - shorter to force faster completion
	ctx, cancelTimeout := context.WithTimeout(ctx, 25*time.Second)
	defer cancelTimeout()

	testFunc(ctx)
}

// LoadTestEnvForInstance loads test environment for a specific instance (convenience wrapper)
func LoadTestEnvForInstance(instance *TestInstance) (alphaSquid, bravoSquid string, err error) {
	return instance.LoadTestEnvForInstance()
}

// GetBaseURLForInstance returns the base URL for a test instance (convenience wrapper)
func GetBaseURLForInstance(instance *TestInstance) string {
	return instance.GetTestInstanceBaseURL()
}

// LoginToAlphaAccount performs a standard login to the Alpha account with ChromeDP
func LoginToAlphaAccount(ctx context.Context, instance *TestInstance) error {
	alphaSquid, _, err := LoadTestEnvForInstance(instance)
	if err != nil {
		return fmt.Errorf("load test environment: %w", err)
	}

	loginVals := url.Values{}
	loginVals.Set("squid", alphaSquid)
	loginVals.Set("account", AlphaAccountName)

	baseURL := GetBaseURLForInstance(instance)
	loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

	return chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(UsernameSelector, AlphaOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(PasswordSelector, TestPassword, chromedp.ByQuery),
		chromedp.Click(LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for login to complete
	)
}

// LoginToBravoAccount performs a standard login to the Bravo account with ChromeDP
func LoginToBravoAccount(ctx context.Context, instance *TestInstance) error {
	_, bravoSquid, err := LoadTestEnvForInstance(instance)
	if err != nil {
		return fmt.Errorf("load test environment: %w", err)
	}

	loginVals := url.Values{}
	loginVals.Set("squid", bravoSquid)
	loginVals.Set("account", BravoAccountName)

	baseURL := GetBaseURLForInstance(instance)
	loginURL := baseURL + "/lemc/login?" + loginVals.Encode()

	return chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(UsernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(UsernameSelector, BravoOwnerUsername, chromedp.ByQuery),
		chromedp.SendKeys(PasswordSelector, TestPassword, chromedp.ByQuery),
		chromedp.Click(LoginButtonSelector, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for login to complete
	)
}
