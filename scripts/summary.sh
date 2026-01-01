#!/bin/bash
# Pipeline summary helper for vibe-dash
# Usage: source scripts/summary.sh; print_test_summary <exit_code> <duration> <output_file>

# ANSI color codes
GREEN='\033[32m'
RED='\033[31m'
RESET='\033[0m'

# Disable colors if not terminal
if [ ! -t 1 ]; then
    GREEN='' RED='' RESET=''
fi

print_separator() {
    echo "════════════════════════════════════════════════════════════════"
}

# Parse test counts from go test -v output
# CRITICAL: Requires -v flag to get individual test results (--- PASS: lines)
# Without -v, only package-level results (ok/FAIL) are available
print_test_summary() {
    local exit_code=$1
    local duration=$2
    local output_file=$3

    # Count package-level results (works with both -v and non-v)
    local passed_pkgs=$(grep -c "^ok\s" "$output_file" 2>/dev/null || echo 0)
    local failed_pkgs=$(grep -c "^FAIL\s" "$output_file" 2>/dev/null || echo 0)

    # Count individual tests (only works with -v flag)
    local passed_tests=$(grep -c "^--- PASS:" "$output_file" 2>/dev/null || echo 0)
    local failed_tests=$(grep -c "^--- FAIL:" "$output_file" 2>/dev/null || echo 0)

    # Use individual test counts if available, otherwise show package counts
    local count_label=""
    if [ "$passed_tests" -gt 0 ] || [ "$failed_tests" -gt 0 ]; then
        count_label="${passed_tests} tests"
    else
        count_label="${passed_pkgs} packages"
    fi

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Tests  ${GREEN}PASS${RESET}  (${duration}s, ${count_label})"
    else
        if [ "$failed_tests" -gt 0 ]; then
            echo -e " ${RED}✗${RESET} Tests  ${RED}FAIL${RESET}  (${passed_tests} passed, ${failed_tests} failed)"
        else
            echo -e " ${RED}✗${RESET} Tests  ${RED}FAIL${RESET}  (${passed_pkgs} ok, ${failed_pkgs} failed packages)"
        fi
    fi
    print_separator

    # Cleanup temp file
    rm -f "$output_file" 2>/dev/null
}

print_lint_summary() {
    local exit_code=$1
    local duration=$2

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Lint   ${GREEN}PASS${RESET}  (${duration}s)"
    else
        echo -e " ${RED}✗${RESET} Lint   ${RED}FAIL${RESET}  (see errors above)"
    fi
    print_separator
}

print_build_summary() {
    local exit_code=$1
    local duration=$2
    local binary_path=$3
    local version=$4

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Build  ${GREEN}PASS${RESET}  (${duration}s, ${binary_path} ${version})"
    else
        echo -e " ${RED}✗${RESET} Build  ${RED}FAIL${RESET}  (see errors above)"
    fi
    print_separator
}
