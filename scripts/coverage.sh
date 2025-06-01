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
    # Get array length and calculate indices
    len=${#files[@]}
    previous=${files[$((len-2))]}
    latest=${files[$((len-1))]}
    
    echo -e "\n📊 === Coverage Comparison ==="
    echo "📋 Previous: $(basename $previous .out)"
    echo "📋 Latest: $(basename $latest .out)"
    
    echo -e "\n📈 === Coverage Totals ==="
    echo -n "⬅️  Previous: "
    go tool cover -func=$previous | tail -1 | awk '{print $3}'
    echo -n "➡️  Latest: "
    go tool cover -func=$latest | tail -1 | awk '{print $3}'
    
    # Calculate improvement (if bc is available)
    if command -v bc >/dev/null 2>&1; then
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
    fi
else
    echo -e "\n📊 === Current Coverage ==="
    go tool cover -func=coverage_${timestamp}.out | tail -1
fi

echo -e "\n✅ Coverage report generated!"
echo "🌐 View HTML report: coverage/coverage_${timestamp}.html"
echo "📄 Raw data: coverage/coverage_${timestamp}.out" 