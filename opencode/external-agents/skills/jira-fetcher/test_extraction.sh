#!/bin/bash

# Test script for ticket ID extraction function

# Copy the extraction function here to avoid calling main
extract_ticket_id() {
    local input="$1"
    
    # If it's already in the format PROJECT-NUMBER
    if [[ "$input" =~ ^[A-Z]+-[0-9]+$ ]]; then
        echo "$input"
        return 0
    fi
    
    # If it's a full URL
    if [[ "$input" =~ https?://[^/]+/browse/([A-Z]+-[0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi
    
    # If it contains the pattern somewhere
    if [[ "$input" =~ ([A-Z]+-[0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi
    
    # If none matched
    echo "Error: Could not extract ticket ID from '$input'" >&2
    return 1
}

echo "Testing ticket ID extraction function..."

# Test cases
test_cases=(
    "CXPVSP-123"
    "https://jira-us-aholddelhaize.atlassian.net/browse/CXPVSP-456"
    "Please look at ticket CXPVSP-789 for details"
    "INVALID-FORMAT"
    "no-ticket-here"
)

expected_results=(
    "CXPVSP-123"
    "CXPVSP-456"
    "CXPVSP-789"
    "Error: Could not extract ticket ID from 'INVALID-FORMAT'"
    "Error: Could not extract ticket ID from 'no-ticket-here'"
)

passed=0
failed=0

for i in "${!test_cases[@]}"; do
    input="${test_cases[$i]}"
    expected="${expected_results[$i]}"
    
    # Capture output and exit code
    output=$(extract_ticket_id "$input" 2>&1)
    exit_code=$?
    
    if [[ "$output" == "$expected" ]]; then
        echo "✓ Test $((i+1)) PASSED: '$input' -> '$output'"
        ((passed++))
    else
        echo "✗ Test $((i+1)) FAILED: '$input'"
        echo "  Expected: '$expected'"
        echo "  Got:      '$output'"
        ((failed++))
    fi
done

echo ""
echo "Results: $passed passed, $failed failed"

if [[ $failed -eq 0 ]]; then
    echo "All tests passed!"
    exit 0
else
    echo "Some tests failed!"
    exit 1
fi