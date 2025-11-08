# Performance Testing with k6

This directory contains performance and load tests using [k6](https://k6.io/).

## Prerequisites

Install k6:

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```powershell
choco install k6
```

Or download from: https://k6.io/docs/getting-started/installation/

## Test Files

### Load Tests

**`load-test.js`** - Standard load testing
- Simulates realistic user behavior
- Gradual ramp-up and ramp-down
- Measures baseline performance
- **Duration:** ~16 minutes
- **Max VUs:** 100

```bash
k6 run load-test.js
```

**Custom options:**
```bash
k6 run --vus 50 --duration 5m load-test.js
BASE_URL=https://your-app.com k6 run load-test.js
```

### Spike Tests

**`spike-test.js`** - Traffic spike testing
- Sudden traffic increases
- Tests auto-scaling
- Validates resilience
- **Duration:** ~10 minutes
- **Max VUs:** 1000

```bash
k6 run spike-test.js
```

### Stress Tests

**`stress-test.js`** - Find breaking point
- Gradually increases load
- Identifies system limits
- Observes degradation patterns
- **Duration:** ~21 minutes
- **Max VUs:** 1000

```bash
k6 run stress-test.js
```

### Scenario Tests

**`scenarios/black-friday.js`** - E-commerce peak traffic
- Simulates Black Friday traffic
- Multiple concurrent scenarios
- Heavy cart and checkout operations
- **Duration:** 2 hours
- **Max VUs:** 1000+

```bash
k6 run scenarios/black-friday.js
```

### API Performance Tests

**`api-performance-test.js`** - Direct API benchmarking
- Tests backend APIs directly
- Constant request rate
- Measures API response times
- **Duration:** 5 minutes
- **Rate:** 150 RPS

```bash
k6 run api-performance-test.js
```

## Running Tests

### Quick Start

```bash
# Run all tests with the helper script
./run_k6_tests.sh

# Or run individual tests
k6 run load-test.js
k6 run --vus 500 --duration 30s spike-test.js
```

### With Environment Variables

```bash
# Test against different environment
BASE_URL=https://staging.myapp.com k6 run load-test.js

# For API tests
PRODUCT_CATALOG_URL=http://localhost:3550 \
RECOMMENDATION_URL=http://localhost:8080 \
k6 run api-performance-test.js
```

### Output Options

**JSON Output:**
```bash
k6 run --out json=results/load-test.json load-test.js
```

**CSV Output:**
```bash
k6 run --out csv=results/load-test.csv load-test.js
```

**InfluxDB (for Grafana):**
```bash
k6 run --out influxdb=http://localhost:8086/k6 load-test.js
```

**Cloud (k6 Cloud):**
```bash
k6 cloud load-test.js
```

## Understanding Results

### Metrics

k6 reports several key metrics:

**HTTP Metrics:**
- `http_req_duration`: Request response time
  - `p(95)`: 95th percentile
  - `p(99)`: 99th percentile
- `http_req_failed`: Percentage of failed requests
- `http_reqs`: Total HTTP requests

**Custom Metrics:**
- `errors`: Custom error rate
- `product_view_duration`: Product page load time
- `checkout_duration`: Checkout completion time

### Thresholds

Tests define thresholds that must pass:

```javascript
thresholds: {
  http_req_duration: ['p(95)<500'],    // 95% of requests under 500ms
  http_req_failed: ['rate<0.05'],       // Less than 5% errors
}
```

**Green checkmark (✓):** Threshold passed
**Red X (✗):** Threshold failed

### Sample Output

```
scenarios: (100.00%) 1 scenario, 100 max VUs, 16m30s max duration
default: 100 looping VUs for 16m0s

✓ homepage status is 200
✓ product page loads in <1s
✗ checkout completes in <2s

checks.........................: 95.23% ✓ 9523     ✗ 477
data_received..................: 45 MB  47 kB/s
data_sent......................: 3.2 MB 3.3 kB/s
http_req_duration..............: avg=234ms min=45ms med=198ms max=3.2s p(95)=456ms
✓ { expected_response:true }...: avg=198ms min=45ms med=176ms max=2.1s p(95)=389ms
http_req_failed................: 4.76%  ✓ 477      ✗ 9523
http_reqs......................: 10000  10.4/s
```

## Performance Benchmarks

### Baseline Targets

For this microservices demo:

| Metric | Target | Acceptable |
|--------|--------|------------|
| Homepage (p95) | < 300ms | < 500ms |
| Product Page (p95) | < 400ms | < 800ms |
| Add to Cart (p95) | < 200ms | < 400ms |
| Checkout (p95) | < 1000ms | < 2000ms |
| Error Rate | < 1% | < 5% |

### Capacity Planning

| Scenario | Concurrent Users | Throughput |
|----------|-----------------|------------|
| Normal Traffic | 100 | 50 RPS |
| Peak Hours | 500 | 250 RPS |
| Black Friday | 1000+ | 500+ RPS |

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Performance Tests

on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install k6
        run: |
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Run load test
        run: |
          cd tests/performance
          k6 run --out json=results/load-test.json load-test.js
        env:
          BASE_URL: ${{ secrets.STAGING_URL }}

      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: k6-results
          path: tests/performance/results/
```

## Visualization

### Option 1: k6 Cloud

```bash
k6 login cloud
k6 cloud load-test.js
```

View results at: https://app.k6.io/

### Option 2: InfluxDB + Grafana

1. Start InfluxDB and Grafana:
```bash
docker-compose -f docker-compose-monitoring.yml up -d
```

2. Run test with InfluxDB output:
```bash
k6 run --out influxdb=http://localhost:8086/k6 load-test.js
```

3. View in Grafana: http://localhost:3000

### Option 3: Custom HTML Reports

Use `k6-reporter`:
```bash
npm install -g k6-to-junit
k6 run --out json=results.json load-test.js
k6-to-junit results.json > results.xml
```

## Troubleshooting

### High Error Rates

**Problem:** `http_req_failed` > 5%

**Solutions:**
- Check if services are running
- Verify BASE_URL is correct
- Reduce VUs or increase ramp-up time
- Check service logs for errors

### Slow Response Times

**Problem:** `http_req_duration` p(95) > 1s

**Solutions:**
- Profile application code
- Check database query performance
- Verify adequate resources (CPU, memory)
- Consider caching strategies

### Connection Errors

**Problem:** `dial: connection refused`

**Solutions:**
- Ensure services are accessible
- Check firewall rules
- Verify correct ports
- Test with `curl` first

### Resource Limits

**Problem:** k6 crashes or hits OS limits

**Solutions:**
- Increase OS file descriptor limit:
  ```bash
  ulimit -n 65536
  ```
- Use distributed k6 execution
- Reduce max VUs

## Best Practices

### 1. Start Small

Always start with low load and gradually increase:
```bash
k6 run --vus 10 --duration 30s load-test.js
```

### 2. Test in Staging First

Never run stress tests against production without approval.

### 3. Monitor During Tests

Watch metrics in real-time:
- CPU usage
- Memory usage
- Database connections
- Error logs

### 4. Clean Up Data

Performance tests can generate lots of data. Clean up:
```bash
# Clear test orders, carts, etc.
./cleanup-test-data.sh
```

### 5. Document Baselines

Record baseline metrics for comparison:
```bash
# Save baseline
k6 run load-test.js | tee baseline-$(date +%Y%m%d).log
```

## Additional Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://k6.io/docs/examples/)
- [Performance Testing Guidance](https://martinfowler.com/articles/practical-test-pyramid.html#PerformanceTests)
- [k6 Community Forum](https://community.k6.io/)
