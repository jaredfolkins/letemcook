# Parallel Test Infrastructure

This document describes the new parallel test infrastructure that allows each test to run with its own isolated resources.

## Overview

The parallel test infrastructure provides:

1. **Unique Test Instances**: Each test gets its own server instance with unique port and data directory
2. **Dynamic Timeouts**: Timeout calculation based on test characteristics and environment
3. **Always Cleanup**: Resources are always cleaned up after test completion (both success and failure)
4. **True Parallelism**: Tests can run simultaneously without resource conflicts

## Quick Start

### For ChromeDP Tests

```go
func TestMyChromeTest(t *testing.T) {
    testutil.ParallelTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
        // Load test environment for this instance
        alphaSquid, _, err := testutil.LoadTestEnvForInstance(instance)
        if err != nil {
            t.Fatalf("Failed to load test environment: %v", err)
        }

        // Use ChromeDP with the instance
        testutil.ChromeDPTestWrapperWithInstance(t, instance, func(ctx context.Context) {
            baseURL := testutil.GetBaseURLForInstance(instance)
            
            // Your test logic here
            tasks := chromedp.Tasks{
                chromedp.Navigate(baseURL + "/lemc/login"),
                // ... more tasks
            }
            
            if err := chromedp.Run(ctx, tasks); err != nil {
                t.Fatalf("ChromeDP tasks failed: %v", err)
            }
        })
    })
}
```

### For Simple HTTP Tests

```go
func TestMyHTTPTest(t *testing.T) {
    t.Parallel() // Enable Go's parallel testing
    
    testutil.SimpleTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
        baseURL := testutil.GetBaseURLForInstance(instance)
        
        client := &http.Client{Timeout: 5 * time.Second}
        resp, err := client.Get(baseURL + "/api/endpoint")
        if err != nil {
            t.Fatalf("HTTP request failed: %v", err)
        }
        defer resp.Body.Close()
        
        // Test response...
    })
}
```

## Test Instance Properties

Each `TestInstance` provides:

- **`ID`**: 8-character lowercase alphanumeric identifier (e.g., "a1b2c3d4")
- **`Port`**: Unique port number (starting from 15362)
- **`DataDir`**: Test-specific data directory (e.g., `/tmp/lemc-data/test-a1b2c3d4`)
- **`TestDataDir`**: Subdirectory for test data (e.g., `/tmp/lemc-data/test-a1b2c3d4/test`)
- **`BaseURL`**: Full base URL (e.g., "http://localhost:15364")

## Dynamic Timeouts

Timeouts are calculated automatically based on:

1. **Test Name Patterns**:
   - `integration`, `e2e`, `full` → 2x base timeout
   - `chrome`, `browser`, `ui` → 3x base timeout  
   - `race`, `concurrent`, `stress` → 2x base timeout

2. **Parallel Execution**: +50% when running in parallel

3. **Environment Configuration**:
   ```bash
   # Base timeout (default: 30s)
   export LEMC_TEST_BASE_TIMEOUT=60s
   
   # Server startup timeout (default: 30s)
   export LEMC_TEST_STARTUP_TIMEOUT=45s
   
   # Override any calculated timeout
   export LEMC_TEST_OVERRIDE_TIMEOUT=5m
   ```

## Always Cleanup

Resources are managed as follows:

- **Test Success**: Server shuts down gracefully, and all resources (data directory, temp files) are cleaned up
- **Test Failure**: Server is killed and all resources (data directory, temp files) are cleaned up  
- **Consistent Behavior**: Resources are always cleaned up regardless of test outcome to prevent accumulation

## Directory Structure

Test instances create the following structure:

```
/tmp/lemc-data/
├── test-a1b2c3d4/          # Instance a1b2c3d4
│   ├── test/               # Test-specific data
│   ├── sessions/           # Session data
│   └── queues/             # Job queues
├── test-x9y8z7w6/          # Instance x9y8z7w6
│   ├── test/
│   ├── sessions/
│   └── queues/
└── ...
```

## Port Allocation

Ports are allocated dynamically starting from 15362:

- Test finds first available port >= 15362
- Each instance gets a unique port
- Supports up to 1000 concurrent test instances (15362-16361)

## Environment Variables

### Test Configuration

```bash
# Timeout configuration
export LEMC_TEST_BASE_TIMEOUT=30s      # Base timeout for tests
export LEMC_TEST_STARTUP_TIMEOUT=30s   # Server startup timeout
export LEMC_TEST_SHUTDOWN_GRACE=5s     # Graceful shutdown time
export LEMC_TEST_OVERRIDE_TIMEOUT=5m   # Override calculated timeout

# Bounds
export LEMC_TEST_MIN_TIMEOUT=10s       # Minimum timeout
export LEMC_TEST_MAX_TIMEOUT=10m       # Maximum timeout
```

### Per-Instance Environment

Each test server runs with:

```bash
LEMC_ENV=test
LEMC_DATA=/tmp/lemc-data/test-{instance-id}
LEMC_SQUID_ALPHABET=abcdefghijklmnopqrstuvwxyz0123456789
LEMC_PORT_TEST={unique-port}
```

## Best Practices

### 1. Use Appropriate Wrapper

- **`ParallelTestWrapper`**: For tests that need full server infrastructure
- **`SimpleTestWrapper`**: For simpler tests that still need a server
- **`ChromeDPTestWrapperWithInstance`**: For browser automation tests

### 2. Enable Go Parallelism

```go
func TestMyTest(t *testing.T) {
    t.Parallel() // Enable Go's parallel execution
    
    testutil.SimpleTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
        // Test logic
    })
}
```

### 3. Use Instance-Specific URLs

Always use the instance-specific base URL:

```go
baseURL := testutil.GetBaseURLForInstance(instance)
loginURL := baseURL + "/lemc/login"
```

### 4. Handle Test Failure Debugging

With always cleanup, debugging requires different approaches:

```bash
# Resources are always cleaned up, so for debugging:
# 1. Add debug logging to your tests
# 2. Use -v flag for verbose output
# 3. Add temporary debug prints in test code
# 4. Use browser dev tools during ChromeDP tests (disable headless mode temporarily)
```

### 5. Configure Timeouts Appropriately

For slow tests or CI environments:

```bash
export LEMC_TEST_BASE_TIMEOUT=60s
export LEMC_TEST_STARTUP_TIMEOUT=60s
```

## Migration from Shared Server Tests

### Before (Shared Server)

```go
func TestOldWay(t *testing.T) {
    // Assumes shared server started in TestMain
    baseURL := testutil.GetBaseURL()
    
    // Test logic using shared resources
}
```

### After (Parallel Instance)

```go
func TestNewWay(t *testing.T) {
    t.Parallel()
    
    testutil.SimpleTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
        baseURL := testutil.GetBaseURLForInstance(instance)
        
        // Same test logic, but with instance-specific resources
    })
}
```

## Troubleshooting

### Port Conflicts

If you see "port already in use" errors:

```bash
# Check what's using ports
lsof -i :15362-15500

# Kill orphaned processes
pkill -f "go run.*letemcook"
```

### Resource Cleanup

Clean up all test resources (this normally happens automatically):

```bash
# Remove any leftover test data directories (shouldn't be needed)
rm -rf /tmp/lemc-data/test-*

# Kill any orphaned test servers (shouldn't be needed)
pkill -f "LEMC_ENV=test"

# Note: The parallel infrastructure automatically cleans up after each test
# These commands are only needed if something goes wrong with the cleanup process
```

### Timeout Issues

Adjust timeouts for your environment:

```bash
# For slower CI environments
export LEMC_TEST_BASE_TIMEOUT=120s
export LEMC_TEST_STARTUP_TIMEOUT=60s

# For faster development
export LEMC_TEST_BASE_TIMEOUT=15s
export LEMC_TEST_STARTUP_TIMEOUT=15s
```

## Performance Considerations

### Parallel Execution

- Tests can run truly in parallel without conflicts
- Each instance uses separate resources (port, database, files)
- Go's `-parallel` flag controls maximum concurrent tests

### Resource Usage

- Each test instance requires ~50-100MB RAM
- Unique port per instance (supports 1000+ concurrent tests)
- Temporary disk space in `/tmp/lemc-data/`

### CI Configuration

For CI environments, consider:

```bash
# Limit parallel tests to avoid resource exhaustion
go test -parallel 4 ./tests/...

# Increase timeouts for slower CI
export LEMC_TEST_BASE_TIMEOUT=2m
export LEMC_TEST_STARTUP_TIMEOUT=90s
``` 