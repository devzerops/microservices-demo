# Contract Testing with Pact

This directory contains contract tests for the microservices using Pact.

## What is Contract Testing?

Contract testing verifies that services can communicate with each other by testing the contracts between consumer and provider services. This allows teams to:

- Deploy services independently
- Catch breaking changes early
- Document API expectations
- Reduce integration test complexity

## Directory Structure

```
contract/
├── consumer/              # Consumer contract tests
│   └── test_recommendation_consumer.py
├── provider/              # Provider verification tests
│   └── test_productcatalog_provider.py
├── pacts/                 # Generated pact files (contracts)
└── requirements.txt       # Dependencies
```

## How It Works

### 1. Consumer Tests (Define Contracts)

Consumer tests define what the consumer expects from the provider:

```python
# Consumer: RecommendationService expects ProductCatalogService to:
(pact
 .given('products exist in catalog')
 .upon_receiving('a request for all products')
 .with_request('get', '/products')
 .will_respond_with(200, body=expected_response))
```

This generates a **pact file** (JSON contract) in `pacts/` directory.

### 2. Provider Tests (Verify Contracts)

Provider tests verify that the actual service satisfies the contracts:

```python
# Provider: ProductCatalogService must satisfy all consumer expectations
verifier.verify_pacts(pact_dir)
```

## Running Contract Tests

### Step 1: Run Consumer Tests

```bash
cd tests/contract

# Install dependencies
pip install -r requirements.txt

# Run consumer tests (generates pact files)
pytest consumer/test_recommendation_consumer.py -v
```

This creates pact files in `pacts/` directory:
- `recommendationservice-productcatalogservice.json`
- `checkoutservice-paymentservice.json`

### Step 2: Run Provider Tests

```bash
# Make sure the provider service is running
cd ../../src/productcatalogservice
go run .

# In another terminal, verify the contracts
cd ../../tests/contract
pytest provider/test_productcatalog_provider.py -v
```

## Contract Examples

### Example 1: Product List Contract

**Consumer Expectation (RecommendationService):**
```json
{
  "products": [
    {
      "id": "OLJCESPC7Z",
      "name": "Sunglasses",
      "price_usd": {
        "currency_code": "USD",
        "units": 19,
        "nanos": 990000000
      }
    }
  ]
}
```

**Provider Must Return:** Same structure with matching types

### Example 2: Payment Charge Contract

**Consumer Expectation (CheckoutService):**
- POST to `/charge` with payment details
- Expects `200 OK` with `transaction_id`

**Provider Must Implement:** Endpoint that returns transaction ID

## Benefits

### For Development Teams

1. **Fast Feedback**: Find breaking changes before integration
2. **Independent Development**: Teams can work without running all services
3. **Living Documentation**: Contracts document API expectations
4. **Confident Refactoring**: Know immediately if changes break contracts

### For CI/CD

1. **Faster Pipelines**: No need to spin up all services
2. **Earlier Detection**: Catch issues before integration tests
3. **Clear Ownership**: Know which team broke the contract

## Best Practices

### 1. Use Provider States

```python
.given('products exist in catalog')  # Sets up provider state
```

Provider should have endpoints to set up these states.

### 2. Use Matchers

```python
Like('Sunglasses')      # Type matching
EachLike({...})         # Array matching
Term(r'^\d+$', '123')   # Regex matching
```

### 3. Keep Contracts Minimal

Only test what the consumer actually uses, not the entire response.

### 4. Version Your Contracts

- Tag pact files with consumer version
- Use Pact Broker for centralized storage

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Contract Tests

on: [push, pull_request]

jobs:
  consumer-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run consumer tests
        run: |
          cd tests/contract
          pip install -r requirements.txt
          pytest consumer/ -v
      - name: Upload pacts
        uses: actions/upload-artifact@v2
        with:
          name: pacts
          path: tests/contract/pacts/

  provider-tests:
    needs: consumer-tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [productcatalogservice, paymentservice]
    steps:
      - uses: actions/checkout@v2
      - name: Download pacts
        uses: actions/download-artifact@v2
        with:
          name: pacts
          path: tests/contract/pacts/
      - name: Start provider service
        run: |
          cd src/${{ matrix.service }}
          go run . &
          sleep 5
      - name: Verify contracts
        run: |
          cd tests/contract
          pip install -r requirements.txt
          pytest provider/ -v -k ${{ matrix.service }}
```

## Advanced: Pact Broker

For production use, consider using [Pact Broker](https://docs.pact.io/pact_broker):

1. **Centralized Storage**: Store all contracts in one place
2. **Webhook Triggers**: Auto-verify when provider changes
3. **Can I Deploy**: Check if it's safe to deploy
4. **Network Diagram**: Visualize service dependencies

```bash
# Publish pacts to broker
pact-broker publish pacts/ \
  --consumer-app-version=1.0.0 \
  --broker-base-url=https://your-pact-broker.com
```

## Troubleshooting

### Pact Verification Fails

1. Check provider service is running
2. Verify provider states are set up correctly
3. Check response structure matches contract
4. Enable verbose logging: `pytest -v --log-cli-level=DEBUG`

### Provider States Not Working

Ensure provider has endpoint to set up states:
```python
@app.route('/_pact/provider_states', methods=['POST'])
def provider_states():
    # Set up database state
    return {'status': 'success'}
```

### Pact Files Not Generated

- Check consumer tests are passing
- Verify `pact_dir` is correct
- Ensure pact.stop() is called (use atexit)

## Resources

- [Pact Documentation](https://docs.pact.io/)
- [Pact Python](https://github.com/pact-foundation/pact-python)
- [Contract Testing Guide](https://martinfowler.com/bliki/ContractTest.html)
- [Pact Best Practices](https://docs.pact.io/getting_started/best_practices)
