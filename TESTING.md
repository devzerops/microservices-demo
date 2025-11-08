# Testing Guide for Microservices Demo

This document provides instructions for running tests across the microservices in this project.

## Overview

The following services now have comprehensive test coverage:

### Python Services
- **emailservice** - Email notification service tests
- **recommendationservice** - Product recommendation service tests

### Go Services
- **checkoutservice** - Checkout service tests (enhanced)
- **shippingservice** - Shipping service tests (enhanced)
- **productcatalogservice** - Product catalog tests (existing)
- **frontend** - Frontend service tests (existing)

---

## Python Services Testing

### Prerequisites

For all Python services, you'll need to install test dependencies:

```bash
pip install pytest pytest-mock pytest-grpc pytest-cov grpcio-testing
```

### Email Service Tests

**Location:** `src/emailservice/`

**Setup:**
```bash
cd src/emailservice

# Install dependencies
pip install -r requirements.txt
pip install -r requirements-test.in

# Run tests
pytest test_email_server.py -v

# Run with coverage
pytest test_email_server.py --cov=email_server --cov-report=html
```

**Test Coverage:**
- BaseEmailService health checks
- DummyEmailService order confirmation
- gRPC integration tests
- Template rendering validation

### Recommendation Service Tests

**Location:** `src/recommendationservice/`

**Setup:**
```bash
cd src/recommendationservice

# Install dependencies
pip install -r requirements.txt
pip install -r requirements-test.in

# Run tests
pytest test_recommendation_server.py -v

# Run with coverage
pytest test_recommendation_server.py --cov=recommendation_server --cov-report=html
```

**Test Coverage:**
- Product recommendation algorithm
- Product filtering logic
- Random selection validation
- Health check endpoints
- gRPC integration tests

---

## Go Services Testing

### Prerequisites

Ensure you have Go 1.23 or later installed:

```bash
go version
```

### Running Tests

All Go services can be tested using the standard Go testing framework:

#### Shipping Service Tests

**Location:** `src/shippingservice/`

```bash
cd src/shippingservice

# Run all tests
go test -v ./...

# Run specific test file
go test -v shippingservice_test.go
go test -v shippingservice_enhanced_test.go

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Coverage:**
- Quote calculation for different cart sizes
- Shipping cost validation
- Tracking ID generation and format validation
- Tracking ID uniqueness
- Multiple address handling
- Edge cases (empty cart, large quantities)
- Health check endpoints

#### Checkout Service Tests

**Location:** `src/checkoutservice/`

```bash
cd src/checkoutservice

# Run all tests
go test -v ./...

# Run specific tests
go test -v checkout_test.go

# Run money operation tests
go test -v money/money_test.go

# Run with coverage
go test -cover ./...
```

**Test Coverage:**
- Money arithmetic operations (addition, multiplication)
- Money validation
- Order ID generation and uniqueness
- Currency handling
- Health check endpoints
- gRPC status code handling

#### Product Catalog Service Tests

**Location:** `src/productcatalogservice/`

```bash
cd src/productcatalogservice

# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

**Test Coverage:**
- Product retrieval
- Product listing
- Product search functionality
- Not found error handling

#### Frontend Service Tests

**Location:** `src/frontend/`

```bash
cd src/frontend

# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

**Test Coverage:**
- Money operations
- Validator functions

---

## Running All Tests

### Python Tests (all services)

```bash
# From repository root
for service in emailservice recommendationservice; do
    echo "Testing $service..."
    cd src/$service
    pip install -q -r requirements.txt -r requirements-test.in
    pytest -v
    cd ../..
done
```

### Go Tests (all services)

```bash
# From repository root
for service in checkoutservice shippingservice productcatalogservice frontend; do
    echo "Testing $service..."
    cd src/$service
    go test -v ./...
    cd ../..
done
```

---

## Test File Structure

### Python Test Files

```
src/emailservice/
├── test_email_server.py        # Main test file
├── requirements-test.in        # Test dependencies
└── pytest.ini                  # Pytest configuration

src/recommendationservice/
├── test_recommendation_server.py  # Main test file
├── requirements-test.in          # Test dependencies
└── pytest.ini                    # Pytest configuration
```

### Go Test Files

```
src/checkoutservice/
├── checkout_test.go           # New comprehensive tests
└── money/
    └── money_test.go          # Money operations tests

src/shippingservice/
├── shippingservice_test.go           # Existing basic tests
└── shippingservice_enhanced_test.go  # Enhanced test coverage

src/productcatalogservice/
└── product_catalog_test.go    # Product catalog tests

src/frontend/
├── money/
│   └── money_test.go          # Money helper tests
└── validator/
    └── validator_test.go      # Validation tests
```

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: Run Tests

on: [push, pull_request]

jobs:
  test-python:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [emailservice, recommendationservice]
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-python@v2
        with:
          python-version: '3.11'
      - name: Install dependencies
        run: |
          cd src/${{ matrix.service }}
          pip install -r requirements.txt -r requirements-test.in
      - name: Run tests
        run: |
          cd src/${{ matrix.service }}
          pytest -v

  test-go:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [checkoutservice, shippingservice, productcatalogservice, frontend]
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.23'
      - name: Run tests
        run: |
          cd src/${{ matrix.service }}
          go test -v ./...
```

---

## Troubleshooting

### Python Tests

**Issue:** `ModuleNotFoundError: No module named 'demo_pb2'`
- **Solution:** Make sure you're in the service directory and protobuf files are generated

**Issue:** Import errors for dependencies
- **Solution:** Install all requirements: `pip install -r requirements.txt`

### Go Tests

**Issue:** `cannot find package`
- **Solution:** Run `go mod download` to download dependencies

**Issue:** Network connectivity errors during test
- **Solution:** Some tests may require internet connectivity for Google Cloud dependencies

---

## Test Coverage Summary

| Service | Language | Test Files | Coverage Areas |
|---------|----------|------------|----------------|
| emailservice | Python | 1 | Health checks, order confirmation, gRPC |
| recommendationservice | Python | 1 | Recommendations, filtering, health checks |
| checkoutservice | Go | 2 | Money ops, order IDs, health checks |
| shippingservice | Go | 2 | Quotes, shipping, tracking IDs |
| productcatalogservice | Go | 1 | Products, search |
| frontend | Go | 2 | Money, validation |

---

## Contributing

When adding new tests:

1. Follow existing test patterns in each service
2. Ensure tests are independent and can run in any order
3. Use descriptive test names that explain what is being tested
4. Include both positive and negative test cases
5. Add integration tests where appropriate
6. Update this documentation with new test coverage

---

## Additional Resources

- [pytest documentation](https://docs.pytest.org/)
- [Go testing package](https://golang.org/pkg/testing/)
- [gRPC testing guide](https://grpc.io/docs/languages/go/basics/#testing)
