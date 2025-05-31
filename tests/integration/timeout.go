package tests

import (
	"context"
	"os"
	"testing"
	"time"
)

// DynamicTestConfig holds configuration for dynamic test timeouts
type DynamicTestConfig struct {
	BaseTimeout    time.Duration
	StartupTimeout time.Duration
	ShutdownGrace  time.Duration
	MaxTimeout     time.Duration
	MinTimeout     time.Duration
}

// DefaultDynamicTestConfig returns sensible defaults
func DefaultDynamicTestConfig() *DynamicTestConfig {
	return &DynamicTestConfig{
		BaseTimeout:    30 * time.Second, // Default test timeout
		StartupTimeout: 30 * time.Second, // Server startup timeout
		ShutdownGrace:  5 * time.Second,  // Graceful shutdown time
		MaxTimeout:     10 * time.Minute, // Maximum allowable timeout
		MinTimeout:     10 * time.Second, // Minimum allowable timeout
	}
}

// LoadDynamicTestConfigFromEnv loads configuration from environment variables
func LoadDynamicTestConfigFromEnv() *DynamicTestConfig {
	config := DefaultDynamicTestConfig()

	if timeoutStr := os.Getenv("LEMC_TEST_BASE_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.BaseTimeout = clampTimeout(timeout, config.MinTimeout, config.MaxTimeout)
		}
	}

	if startupStr := os.Getenv("LEMC_TEST_STARTUP_TIMEOUT"); startupStr != "" {
		if timeout, err := time.ParseDuration(startupStr); err == nil {
			config.StartupTimeout = clampTimeout(timeout, config.MinTimeout, config.MaxTimeout)
		}
	}

	if graceStr := os.Getenv("LEMC_TEST_SHUTDOWN_GRACE"); graceStr != "" {
		if timeout, err := time.ParseDuration(graceStr); err == nil {
			config.ShutdownGrace = clampTimeout(timeout, time.Second, 30*time.Second)
		}
	}

	return config
}

// clampTimeout ensures timeout is within reasonable bounds
func clampTimeout(timeout, min, max time.Duration) time.Duration {
	if timeout < min {
		return min
	}
	if timeout > max {
		return max
	}
	return timeout
}

// CalculateDynamicTimeout calculates timeout based on test characteristics
func (config *DynamicTestConfig) CalculateDynamicTimeout(testName string, parallel bool) time.Duration {
	timeout := config.BaseTimeout

	// Adjust based on test name patterns
	if containsAny(testName, []string{"integration", "e2e", "full"}) {
		timeout = timeout * 2
	}
	if containsAny(testName, []string{"chrome", "browser", "ui"}) {
		timeout = timeout * 3
	}
	if containsAny(testName, []string{"race", "concurrent", "stress"}) {
		timeout = timeout * 2
	}

	// Parallel tests get more time due to resource contention
	if parallel {
		timeout = timeout + (timeout / 2) // +50%
	}

	// Respect environment variable override
	if envTimeoutStr := os.Getenv("LEMC_TEST_OVERRIDE_TIMEOUT"); envTimeoutStr != "" {
		if envTimeout, err := time.ParseDuration(envTimeoutStr); err == nil {
			timeout = envTimeout
		}
	}

	return clampTimeout(timeout, config.MinTimeout, config.MaxTimeout)
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// SimpleTestWrapper for non-parallel tests with dynamic timeout and always cleanup
func SimpleTestWrapper(t *testing.T, testFunc func(*testing.T, *TestInstance)) {
	config := LoadDynamicTestConfigFromEnv()

	// Calculate dynamic timeout based on test characteristics
	timeout := config.CalculateDynamicTimeout(t.Name(), false)

	t.Logf("Test %s: calculated timeout %v", t.Name(), timeout)

	// Create test instance with unique resources
	instance, err := CreateTestInstance()
	if err != nil {
		t.Fatalf("Failed to create test instance: %v", err)
	}

	t.Logf("Test %s: created instance %s on port %d", t.Name(), instance.ID, instance.Port)

	// Start server with startup timeout
	startupCtx, startupCancel := context.WithTimeout(context.Background(), config.StartupTimeout)
	defer startupCancel()

	serverCleanup, err := instance.StartTestServer()
	if err != nil {
		serverCleanup(true) // Cleanup on startup failure
		t.Fatalf("Failed to start test server: %v", err)
	}

	// Track test success/failure
	testFailed := false
	defer func() {
		if r := recover(); r != nil {
			testFailed = true
			t.Errorf("Test panicked: %v", r)
		}

		t.Logf("Test %s: cleaning up (failed=%v)", t.Name(), testFailed)
		serverCleanup(testFailed) // Always cleanup regardless of success/failure

		// Force cleanup any remaining ChromeDP contexts
		ForceCleanupChrome()

		// Give a moment for cleanup to complete
		time.Sleep(200 * time.Millisecond)
	}()

	// Create test timeout context
	testCtx, testCancel := context.WithTimeout(startupCtx, timeout)
	defer testCancel()

	// Run test in goroutine to handle timeout
	done := make(chan bool, 1)
	testGoroutine := make(chan struct{})

	go func() {
		defer func() {
			close(testGoroutine) // Signal goroutine is done
			if r := recover(); r != nil {
				testFailed = true
				t.Errorf("Test function panicked: %v", r)
			}
			done <- true
		}()

		// Run the actual test
		testFunc(t, instance)
	}()

	// Wait for test completion or timeout
	select {
	case <-done:
		// Test completed normally
		if t.Failed() {
			testFailed = true
		}
		t.Logf("Test %s: completed (failed=%v)", t.Name(), testFailed)
	case <-testCtx.Done():
		testFailed = true
		t.Errorf("Test %s: timeout after %v", t.Name(), timeout)
	}

	// Ensure test goroutine is properly terminated
	select {
	case <-testGoroutine:
		// Goroutine finished normally
	case <-time.After(100 * time.Millisecond):
		// Goroutine didn't finish, but we'll continue anyway
		t.Logf("Test %s: goroutine cleanup timeout", t.Name())
	}
}

// SeriesTestWrapper for series tests (non-parallel) with dynamic timeout and always cleanup
func SeriesTestWrapper(t *testing.T, testFunc func(*testing.T, *TestInstance)) {
	// Just call SimpleTestWrapper since they're essentially the same for series tests
	SimpleTestWrapper(t, testFunc)
}
