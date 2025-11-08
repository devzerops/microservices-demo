#!/bin/bash

# K6 Performance Test Runner
# Runs various k6 test scenarios and generates reports

set -e

echo "========================================"
echo "K6 Performance Test Suite"
echo "========================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}k6 is not installed!${NC}"
    echo "Install k6: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Base URL
BASE_URL=${BASE_URL:-http://localhost:8080}
echo -e "${BLUE}Testing against: $BASE_URL${NC}\n"

# Create results directory
mkdir -p results

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    local description=$3

    echo -e "${YELLOW}Running: $test_name${NC}"
    echo "$description"
    echo "----------------------------------------"

    if k6 run \
        --out json=results/${test_name}-results.json \
        --summary-export=results/${test_name}-summary.json \
        -e BASE_URL=$BASE_URL \
        $test_file; then
        echo -e "${GREEN}✓ $test_name completed${NC}\n"
        return 0
    else
        echo -e "${RED}✗ $test_name failed${NC}\n"
        return 1
    fi
}

# Test selection
if [ -z "$1" ]; then
    echo "Select test to run:"
    echo "  1) Load Test (recommended first)"
    echo "  2) Spike Test"
    echo "  3) Stress Test"
    echo "  4) Black Friday Scenario"
    echo "  5) API Performance Test"
    echo "  6) All Tests (sequential)"
    echo "  7) Quick Smoke Test"
    read -p "Enter choice (1-7): " choice
else
    choice=$1
fi

case $choice in
    1)
        run_test "load-test" "load-test.js" "Standard load test with gradual ramp-up"
        ;;
    2)
        run_test "spike-test" "spike-test.js" "Sudden traffic spike simulation"
        ;;
    3)
        run_test "stress-test" "stress-test.js" "Find system breaking point"
        ;;
    4)
        run_test "black-friday" "scenarios/black-friday.js" "Black Friday traffic pattern"
        ;;
    5)
        run_test "api-performance" "api-performance-test.js" "Direct API performance testing"
        ;;
    6)
        echo -e "${YELLOW}Running all tests sequentially...${NC}\n"
        run_test "load-test" "load-test.js" "Standard load test"
        sleep 30  # Cool down period
        run_test "spike-test" "spike-test.js" "Spike test"
        sleep 30
        run_test "api-performance" "api-performance-test.js" "API performance"
        ;;
    7)
        echo -e "${YELLOW}Running quick smoke test...${NC}\n"
        k6 run --vus 10 --duration 30s -e BASE_URL=$BASE_URL load-test.js
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

# Generate combined report if multiple tests ran
if [ "$choice" == "6" ]; then
    echo -e "\n${BLUE}Generating combined report...${NC}"
    echo "Results saved in: tests/performance/results/"

    # You could add additional report generation here
    # e.g., using k6-reporter or custom scripts
fi

echo "========================================"
echo -e "${GREEN}Performance testing complete!${NC}"
echo "========================================"
echo "Results location: tests/performance/results/"
echo ""
echo "To visualize results:"
echo "  - Upload JSON results to https://app.k6.io/"
echo "  - Or use k6-to-junit to generate JUnit XML"
echo "  - Or use custom dashboards with InfluxDB/Grafana"
