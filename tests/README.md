# Consolidated Tests

This directory contains all tests for the letemcook application, consolidated from the previous separate `tests/chromedp`, `tests/integration`, and `tests` directories.

## Structure

All tests are now in a single `tests` package with the following components:

### Core Test Infrastructure
- `main_test.go` - TestMain function with comprehensive cleanup
- `timeout.go` - Dynamic timeout configuration and test wrappers
- `parallel.go` - Test instance management for parallel execution
- `process_registry.go` - Process tracking for safe cleanup
- `chromedp.go` - ChromeDP browser automation utilities
- `util.go` - Utility functions (RepoRoot, DataRoot)

### Test Files
- `login_test.go` - Login functionality tests
- `chromedp_integration_test.go` - ChromeDP-specific integration tests
- `app_create_test.go` - App creation tests
- `app_visibility_test.go` - App visibility and permissions tests
- `cookbook_create_test.go` - Cookbook creation tests
- `nav_active_test.go` - Navigation state tests
- `nav_refresh_error_test.go` - Navigation refresh error tests
- `job_status_test.go` - Job status and polling tests
- `cleanup_demo_test.go` - Cleanup demonstration test

## Key Changes

1. **Single Package**: All code is now in the `tests` package instead of separate packages
2. **No Import Dependencies**: Removed all `tests` imports since everything is in the same package
3. **Consolidated Functions**: All utility functions are directly accessible without package prefixes
4. **Fixed Path Resolution**: Updated `RepoRoot()` function to work with the new directory structure
5. **Unified TestMain**: Single TestMain function handles all cleanup and initialization

## Running Tests

```bash
# Run all tests
go test -v

# Run tests in short mode (skips integration tests)
go test -v -short

# Run specific test
go test -v -run TestLogin

# Run with timeout override
LEMC_TEST_OVERRIDE_TIMEOUT=60s go test -v -run TestLogin
```

## Test Instance Management

Each test gets its own isolated test instance with:
- Unique port (starting from 15362)
- Isolated data directory
- Separate server process
- Automatic cleanup on completion

This ensures tests can run in parallel without interference. 