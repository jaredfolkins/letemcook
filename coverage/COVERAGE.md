# Code Coverage Tracking Strategy

This directory contains our code coverage tracking system with timestamped files for progress monitoring.

## File Management Strategy

### Coverage File Format
- **Coverage profiles**: `coverage_<unix_timestamp>.out`
- **HTML reports**: `coverage_<unix_timestamp>.html`
- **Unix timestamps**: Ensure chronological ordering and unique filenames

### Retention Policy
- **Keep maximum 2 sets** of coverage files (latest + previous)
- **Automatic cleanup**: When generating new coverage, remove oldest files if more than 2 sets exist
- **Comparison ready**: Always have previous run available for diff analysis

## Usage Workflow

### 1. Generate New Coverage
```bash
# Generate timestamped coverage files
timestamp=$(date +%s)
go test -coverprofile=coverage/coverage_${timestamp}.out ./...
go tool cover -html=coverage/coverage_${timestamp}.out -o coverage/coverage_${timestamp}.html
```

### 2. Automatic Cleanup (if needed)
```bash
# Keep only the 2 most recent coverage file sets
cd coverage
# List all .out files by timestamp (oldest first)
files=($(ls coverage_*.out 2>/dev/null | sort))
if [ ${#files[@]} -gt 2 ]; then
    # Remove oldest files (both .out and .html)
    oldest=${files[0]}
    basename=${oldest%.*}  # Remove .out extension
    rm -f "${basename}.out" "${basename}.html"
    echo "Removed oldest coverage files: ${basename}.*"
fi
```

### 3. Compare Coverage Progress
```bash
# Get the two most recent coverage files
files=($(ls coverage_*.out 2>/dev/null | sort))
if [ ${#files[@]} -ge 2 ]; then
    previous=${files[-2]}  # Second to last
    latest=${files[-1]}    # Most recent
    
    echo "=== Coverage Comparison ==="
    echo "Previous: $previous"
    echo "Latest: $latest"
    
    # Compare function-level coverage
    echo -e "\n=== Previous Coverage ==="
    go tool cover -func=$previous | tail -1
    
    echo -e "\n=== Latest Coverage ==="
    go tool cover -func=$latest | tail -1
    
    # Optional: Detailed diff of function coverage
    echo -e "\n=== Detailed Function Coverage Diff ==="
    diff <(go tool cover -func=$previous) <(go tool cover -func=$latest) || true
fi
```

## Automation Script

Create a `scripts/coverage.sh` script to automate the entire workflow:

```bash
#!/bin/bash
set -e

echo "🧪 Generating new coverage report..."
timestamp=$(date +%s)
echo "📅 Timestamp: $timestamp"

# Generate coverage
go test -coverprofile=coverage/coverage_${timestamp}.out ./...
go tool cover -html=coverage/coverage_${timestamp}.out -o coverage/coverage_${timestamp}.html

# Cleanup old files (keep only 2 most recent sets)
cd coverage
files=($(ls coverage_*.out 2>/dev/null | sort))
if [ ${#files[@]} -gt 2 ]; then
    # Remove oldest files
    oldest=${files[0]}
    basename=${oldest%.*}
    rm -f "${basename}.out" "${basename}.html"
    echo "🗑️  Removed oldest coverage files: ${basename}.*"
fi

# Show comparison if we have previous data
files=($(ls coverage_*.out 2>/dev/null | sort))
if [ ${#files[@]} -ge 2 ]; then
    previous=${files[-2]}
    latest=${files[-1]}
    
    echo -e "\n📊 === Coverage Comparison ==="
    echo "📋 Previous: $(basename $previous .out)"
    echo "📋 Latest: $(basename $latest .out)"
    
    echo -e "\n📈 === Coverage Totals ==="
    echo -n "⬅️  Previous: "
    go tool cover -func=$previous | tail -1 | awk '{print $3}'
    echo -n "➡️  Latest: "
    go tool cover -func=$latest | tail -1 | awk '{print $3}'
    
    # Calculate improvement
    prev_pct=$(go tool cover -func=$previous | tail -1 | awk '{print $3}' | sed 's/%//')
    latest_pct=$(go tool cover -func=$latest | tail -1 | awk '{print $3}' | sed 's/%//')
    improvement=$(echo "$latest_pct - $prev_pct" | bc -l)
    
    if (( $(echo "$improvement > 0" | bc -l) )); then
        echo "📈 Improvement: +${improvement}%"
    elif (( $(echo "$improvement < 0" | bc -l) )); then
        echo "📉 Decrease: ${improvement}%"
    else
        echo "➡️  No change"
    fi
else
    echo -e "\n📊 === Current Coverage ==="
    go tool cover -func=coverage_${timestamp}.out | tail -1
fi

echo -e "\n✅ Coverage report generated!"
echo "🌐 View HTML report: coverage/coverage_${timestamp}.html"
echo "📄 Raw data: coverage/coverage_${timestamp}.out"
```

## Benefits

1. **Historical Tracking**: Keep previous run for comparison
2. **Automatic Cleanup**: No manual file management needed
3. **Progress Monitoring**: Easy to see if coverage is improving
4. **Unique Filenames**: Unix timestamps prevent conflicts
5. **Git Friendly**: Only .gitkeep is tracked, coverage files are local

## Integration with CI/CD

- Add coverage generation to your CI pipeline
- Store coverage artifacts for trend analysis
- Set coverage thresholds and fail builds if coverage drops
- Generate coverage badges for README

## Files in this Directory

- `.gitkeep`: Ensures directory is tracked in git
- `coverage_<timestamp>.out`: Go coverage profile data
- `coverage_<timestamp>.html`: Visual HTML coverage report
- `COVERAGE.md`: This documentation file 