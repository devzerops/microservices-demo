#!/bin/bash

# Integration Test Runner Script
# This script sets up the test environment and runs integration tests

set -e

echo "========================================"
echo "Integration Test Runner"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Change to tests/integration directory
cd "$(dirname "$0")"

echo -e "${YELLOW}Step 1: Installing Python dependencies...${NC}"
pip install -q -r requirements.txt

echo -e "${YELLOW}Step 2: Starting services with Docker Compose...${NC}"
docker-compose -f docker-compose.test.yml up -d

echo -e "${YELLOW}Step 3: Waiting for services to be healthy...${NC}"
sleep 10

# Check if services are up
echo "Checking service health..."
for i in {1..30}; do
    if docker-compose -f docker-compose.test.yml ps | grep -q "unhealthy"; then
        echo "Waiting for services... ($i/30)"
        sleep 2
    else
        echo -e "${GREEN}All services are healthy!${NC}"
        break
    fi

    if [ $i -eq 30 ]; then
        echo -e "${RED}Services failed to become healthy${NC}"
        docker-compose -f docker-compose.test.yml logs
        docker-compose -f docker-compose.test.yml down
        exit 1
    fi
done

echo -e "${YELLOW}Step 4: Running integration tests...${NC}"

# Set environment variables for test
export PRODUCT_CATALOG_SERVICE_ADDR=localhost:3550
export RECOMMENDATION_SERVICE_ADDR=localhost:8080
export CART_SERVICE_ADDR=localhost:7070
export CHECKOUT_SERVICE_ADDR=localhost:5050

# Run tests
if pytest test_service_integration.py -v --tb=short --timeout=60; then
    echo -e "${GREEN}✓ Integration tests passed!${NC}"
    TEST_RESULT=0
else
    echo -e "${RED}✗ Integration tests failed!${NC}"
    TEST_RESULT=1
fi

echo -e "${YELLOW}Step 5: Cleaning up...${NC}"
docker-compose -f docker-compose.test.yml down -v

echo "========================================"
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}Integration tests completed successfully!${NC}"
else
    echo -e "${RED}Integration tests failed!${NC}"
fi
echo "========================================"

exit $TEST_RESULT
