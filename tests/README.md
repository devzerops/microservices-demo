# Comprehensive Testing Suite

This directory contains a complete testing framework for the microservices demo project.

## 📁 Directory Structure

```
tests/
├── integration/          # Integration tests (service-to-service)
├── contract/            # Contract tests with Pact
├── performance/         # Performance and load tests with k6
└── README.md           # This file
```

## 🎯 Test Types

### 1. Integration Tests (`integration/`)

Tests that verify multiple services working together.

**Technologies:** Python, pytest, gRPC, Docker Compose

**What it tests:**
- Complete business workflows (browse → cart → checkout)
- Service-to-service communication
- Data flow across services
- Health checks for all services
- Error handling and edge cases

**Quick start:**
```bash
cd tests/integration
./run_tests.sh
```

**Key features:**
- ✅ Isolated Docker environment
- ✅ Real service communication
- ✅ Automated setup and teardown
- ✅ Complete checkout flow testing

[📖 Full Integration Testing Guide →](integration/README.md)

---

### 2. Contract Testing (`contract/`)

Consumer-Driven Contract Testing with Pact.

**Technologies:** Pact, Python

**What it tests:**
- API contracts between services
- Consumer expectations vs Provider implementation
- API compatibility and versioning
- Breaking changes detection

**Quick start:**
```bash
cd tests/contract

# Run consumer tests (generates contracts)
pytest consumer/ -v

# Run provider tests (verifies contracts)
pytest provider/ -v
```

**Key features:**
- ✅ Independent service deployment
- ✅ Early breaking change detection
- ✅ Living API documentation
- ✅ Fast feedback loop

[📖 Full Contract Testing Guide →](contract/README.md)

---

### 3. Performance Testing (`performance/`)

Load, stress, and spike testing with k6.

**Technologies:** k6 (JavaScript)

**What it tests:**
- System performance under load
- Response time percentiles (p95, p99)
- Throughput and capacity limits
- System breaking points
- Auto-scaling behavior

**Test scenarios:**
1. **Load Test** - Baseline performance (16 min, 100 VUs)
2. **Spike Test** - Traffic spikes (10 min, up to 1000 VUs)
3. **Stress Test** - Find limits (21 min, up to 1000 VUs)
4. **Black Friday** - E-commerce peak (2 hours, 1000+ VUs)
5. **API Performance** - Backend benchmarking (5 min, 150 RPS)

**Quick start:**
```bash
cd tests/performance
./run_k6_tests.sh
```

**Key features:**
- ✅ Realistic user scenarios
- ✅ Custom metrics and thresholds
- ✅ Multiple test patterns
- ✅ Results visualization

[📖 Full Performance Testing Guide →](performance/README.md)

---

## 🚀 Quick Start Guide

### Prerequisites

**For Integration Tests:**
```bash
# Docker and Docker Compose
docker --version
docker-compose --version

# Python 3.11+
python --version
pip install pytest grpcio grpcio-health-checking
```

**For Contract Tests:**
```bash
pip install pytest pact-python
```

**For Performance Tests:**
```bash
# Install k6
# macOS
brew install k6

# Linux
sudo apt-get install k6

# Or see: https://k6.io/docs/getting-started/installation/
```

### Running All Tests

**Option 1: Run each test suite individually**

```bash
# Integration tests (30-40 minutes)
cd tests/integration
./run_tests.sh

# Contract tests (5 minutes)
cd tests/contract
pytest consumer/ -v
pytest provider/ -v

# Performance tests (choose scenario)
cd tests/performance
./run_k6_tests.sh
```

**Option 2: Automated CI/CD pipeline**

See `.github/workflows/` for CI/CD integration examples.

---

## 📊 Test Matrix

| Test Type | Duration | Complexity | Isolation | Feedback Speed |
|-----------|----------|------------|-----------|----------------|
| Unit Tests | Seconds | Low | High | Very Fast |
| Contract Tests | Minutes | Medium | High | Fast |
| Integration Tests | 30+ min | High | Medium | Medium |
| Performance Tests | Hours | High | Low | Slow |

---

## 🎓 Testing Strategy

### Pyramid Approach

```
         /\
        /  \  E2E/Performance (Few)
       /    \
      /------\
     / Integ  \ Integration Tests (Some)
    /----------\
   /  Contract  \ Contract Tests (More)
  /--------------\
 /   Unit Tests   \ Unit Tests (Many)
/------------------\
```

### When to Use Each Test Type

**Unit Tests** (See `src/*/test_*.py` and `src/*/test_*.go`)
- Testing individual functions
- Fast feedback during development
- High code coverage

**Contract Tests** (`tests/contract/`)
- Defining API contracts
- Preventing breaking changes
- Independent service deployment

**Integration Tests** (`tests/integration/`)
- Verifying service interactions
- Testing complete workflows
- Pre-deployment validation

**Performance Tests** (`tests/performance/`)
- Capacity planning
- Identifying bottlenecks
- SLA validation

---

## 🔄 CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    # ... run unit tests

  contract-tests:
    runs-on: ubuntu-latest
    # ... run contract tests

  integration-tests:
    runs-on: ubuntu-latest
    needs: [unit-tests, contract-tests]
    # ... run integration tests

  performance-tests:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    # ... run performance tests (main branch only)
```

### Test Gates

**Pull Request Requirements:**
- ✅ All unit tests pass
- ✅ Contract tests pass
- ✅ Code coverage > 80%

**Pre-Deployment:**
- ✅ Integration tests pass
- ✅ Performance benchmarks met

**Post-Deployment:**
- ✅ Smoke tests pass
- ✅ Performance monitoring

---

## 📈 Metrics and Reporting

### Integration Tests

**Output:**
- Test results (pass/fail)
- Service logs
- Docker container status

**Example:**
```
test_complete_checkout_flow PASSED [100%]
✓ Integration tests passed!
```

### Contract Tests

**Output:**
- Pact files (JSON contracts)
- Verification results
- Contract diff

**Files generated:**
```
pacts/
├── recommendationservice-productcatalogservice.json
└── checkoutservice-paymentservice.json
```

### Performance Tests

**Metrics:**
- Request duration (avg, p95, p99)
- Throughput (requests/sec)
- Error rate
- Data transferred

**Output formats:**
- Console summary
- JSON results
- CSV export
- Cloud dashboards

**Example:**
```
✓ http_req_duration..........: avg=234ms p(95)=456ms
✓ http_req_failed............: 4.76%
  http_reqs..................: 10000 (10.4/s)
```

---

## 🐛 Debugging Failed Tests

### Integration Tests

**Issue:** Services fail to start
```bash
# Check logs
docker-compose -f tests/integration/docker-compose.test.yml logs

# Check individual service
docker-compose -f tests/integration/docker-compose.test.yml logs productcatalogservice
```

**Issue:** Tests timeout
- Increase timeout in pytest: `pytest --timeout=120`
- Check service health: `docker-compose ps`

### Contract Tests

**Issue:** Contract verification fails
- Check provider is running on correct port
- Verify provider states are implemented
- Compare expected vs actual response structure

### Performance Tests

**Issue:** High error rates
- Reduce virtual users (VUs)
- Increase ramp-up duration
- Check service capacity

**Issue:** Slow response times
- Profile application code
- Check database performance
- Review resource allocation

---

## 🏆 Best Practices

### General

1. **Test Isolation:** Each test should be independent
2. **Clean Data:** Clean up test data after each run
3. **Deterministic:** Tests should give consistent results
4. **Fast Feedback:** Optimize test execution time
5. **Meaningful Names:** Clear test names describing what is tested

### Integration Tests

```python
# Good
def test_complete_checkout_flow_creates_order():
    """Test that completing checkout creates an order with items"""
    # ...

# Bad
def test_checkout():
    # ...
```

### Contract Tests

```python
# Good - Test what consumer actually uses
.with_request('get', '/products')
.will_respond_with(200, body={'products': EachLike({...})})

# Bad - Testing too much
.will_respond_with(200, body={'products': [...], 'metadata': {...}, ...})
```

### Performance Tests

```javascript
// Good - Realistic user behavior
export default function() {
  browseProducts();
  sleep(Math.random() * 3 + 1);  // Random think time
  addToCart();
  sleep(2);
}

// Bad - Unrealistic hammering
export default function() {
  http.get('/');
  http.get('/product/1');
  http.get('/cart');
  // No sleep, no realistic patterns
}
```

---

## 📚 Additional Resources

### Documentation
- [Integration Testing README](integration/README.md)
- [Contract Testing README](contract/README.md)
- [Performance Testing README](performance/README.md)
- [Testing Enhancements Proposal](../TESTING_ENHANCEMENTS_PROPOSAL.md)

### External Resources
- [Martin Fowler - Microservices Testing](https://martinfowler.com/articles/microservice-testing/)
- [Pact Documentation](https://docs.pact.io/)
- [k6 Documentation](https://k6.io/docs/)
- [Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)

### Tools
- [pytest](https://docs.pytest.org/) - Python testing framework
- [Pact](https://docs.pact.io/) - Contract testing
- [k6](https://k6.io/) - Performance testing
- [Docker Compose](https://docs.docker.com/compose/) - Service orchestration

---

## 🤝 Contributing

### Adding New Tests

1. **Integration Tests:**
   - Add test functions to `tests/integration/test_service_integration.py`
   - Update `docker-compose.test.yml` if new services needed

2. **Contract Tests:**
   - Consumer tests in `tests/contract/consumer/`
   - Provider tests in `tests/contract/provider/`

3. **Performance Tests:**
   - Add k6 scripts to `tests/performance/`
   - Update `run_k6_tests.sh` if needed

### Test Review Checklist

- [ ] Tests are independent and isolated
- [ ] Tests have clear, descriptive names
- [ ] Tests clean up after themselves
- [ ] Tests are documented
- [ ] Tests run in CI/CD pipeline
- [ ] Tests have appropriate timeouts
- [ ] Tests handle errors gracefully

---

## 🎯 Test Coverage Goals

| Service | Unit | Integration | Contract | Performance |
|---------|------|-------------|----------|-------------|
| emailservice | ✅ 80% | ✅ | ⬜ | ⬜ |
| recommendationservice | ✅ 75% | ✅ | ✅ | ✅ |
| checkoutservice | ✅ 70% | ✅ | ✅ | ✅ |
| shippingservice | ✅ 85% | ✅ | ⬜ | ✅ |
| productcatalogservice | ✅ 80% | ✅ | ✅ | ✅ |
| cartservice | ⬜ 60% | ✅ | ⬜ | ⬜ |
| frontend | ✅ 65% | ✅ | ⬜ | ✅ |

**Legend:** ✅ Implemented | ⬜ Planned | Percentage = Code Coverage

---

## 📞 Support

For questions or issues:
1. Check the relevant README in each test directory
2. Review [TESTING_ENHANCEMENTS_PROPOSAL.md](../TESTING_ENHANCEMENTS_PROPOSAL.md)
3. Open an issue in the repository

---

**Happy Testing! 🎉**
