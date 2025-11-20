# Environment Variables Reference

This document describes all environment variables used by the Online Boutique microservices, including the enhanced security features.

## Security Features Environment Variables

### Frontend Service

#### Rate Limiting
- `ENABLE_RATE_LIMITING` - Enable/disable rate limiting (default: `false`, set via Helm: `true`)
  - Values: `true`, `1` (enabled) or `false`, empty (disabled)
- `RATE_LIMIT_AI` - AI assistant requests per minute (default: `10`, Helm default: `20`)
  - Range: 1-10000
- `RATE_LIMIT_POST` - POST requests per minute (default: `60`, Helm default: `100`)
  - Range: 1-10000
- `RATE_LIMIT_GET` - GET requests per minute (default: `120`, Helm default: `200`)
  - Range: 1-10000

#### HTTP Security
- `ENABLE_HTTPS` - Enable HSTS header (default: `false`)
  - Values: `true`, `1` (enabled) or `false`, empty (disabled)
  - Only enable when using HTTPS in production

### gRPC TLS Configuration

Applied to: Frontend, Checkoutservice, Recommendationservice

- `ENABLE_GRPC_TLS` - Enable gRPC TLS encryption (default: `false`)
  - Values:
    - `true` or `system`: Use system CA certificates (production)
    - `skip-verify`: Use TLS but skip certificate verification (testing only)
    - `custom`: Use custom CA certificate from file
    - `false` or empty: Insecure connection (demo/development)

- `GRPC_TLS_CA_CERT` - Path to custom CA certificate file (required when `ENABLE_GRPC_TLS=custom`)
  - Example: `/etc/ssl/certs/ca-certificates.crt`

### Database Configuration

#### AlloyDB (CartService)
- `ALLOYDB_USER` - Database user (default: `postgres`)
  - Recommendation: Use dedicated user with least privilege
- `ALLOYDB_SECRET_NAME` - Secret Manager secret name for password
- `ALLOYDB_DATABASE_NAME` - Database name
- `ALLOYDB_TABLE_NAME` - Table name for cart data
- `ALLOYDB_PRIMARY_IP` - Primary instance IP for writes
- `ALLOYDB_READ_IP` - Read pool IP for reads (optional, improves performance)
- `PROJECT_ID` - Google Cloud project ID

#### Spanner (CartService)
- `SPANNER_PROJECT` - Google Cloud project ID
- `SPANNER_INSTANCE` - Spanner instance name
- `SPANNER_DATABASE` - Spanner database name
- `SPANNER_CONNECTION_STRING` - Full connection string (alternative to separate params)

#### Redis (CartService)
- `REDIS_ADDR` - Redis server address (default connects with SSL enabled)

### AI Services (ShoppingAssistantService)

- `PROJECT_ID` - Google Cloud project ID
- `REGION` - Google Cloud region
- `ALLOYDB_DATABASE_NAME` - Vector store database name
- `ALLOYDB_TABLE_NAME` - Vector store table name
- `ALLOYDB_CLUSTER_NAME` - AlloyDB cluster name
- `ALLOYDB_INSTANCE_NAME` - AlloyDB instance name
- `ALLOYDB_SECRET_NAME` - Secret Manager secret for database password
- `PORT` - Service port (default: `8080`)

## Standard Service Environment Variables

### Common Variables

Applied to most services:

- `PORT` - Service port (validated: 1-65535)
- `ENABLE_TRACING` - Enable OpenTelemetry tracing (default: `false`)
  - Values: `1` (enabled) or `0`, empty (disabled)
- `ENABLE_PROFILER` - Enable Cloud Profiler (default: `false`)
  - Values: `1` (enabled) or `0`, empty (disabled)
- `DISABLE_PROFILER` - Disable Cloud Profiler (recommendationservice)
  - Values: `1` (disabled)

### Service Address Variables

Frontend service connects to all other services via these variables:

- `PRODUCT_CATALOG_SERVICE_ADDR` - ProductCatalog service address
- `CURRENCY_SERVICE_ADDR` - Currency service address
- `CART_SERVICE_ADDR` - Cart service address
- `RECOMMENDATION_SERVICE_ADDR` - Recommendation service address
- `SHIPPING_SERVICE_ADDR` - Shipping service address
- `CHECKOUT_SERVICE_ADDR` - Checkout service address
- `AD_SERVICE_ADDR` - Ad service address
- `SHOPPING_ASSISTANT_SERVICE_ADDR` - AI assistant service address
- `COLLECTOR_SERVICE_ADDR` - OpenTelemetry collector address

Checkoutservice connects to:

- `PRODUCT_CATALOG_SERVICE_ADDR`
- `SHIPPING_SERVICE_ADDR`
- `PAYMENT_SERVICE_ADDR`
- `EMAIL_SERVICE_ADDR`
- `CURRENCY_SERVICE_ADDR`
- `CART_SERVICE_ADDR`

Recommendationservice connects to:

- `PRODUCT_CATALOG_SERVICE_ADDR`

### Frontend-Specific Variables

- `LISTEN_ADDR` - Listen address (default: empty, listens on all interfaces)
- `BASE_URL` - Base URL path (default: empty)
- `ENV_PLATFORM` - Platform identifier (values: `local`, `gcp`, `aws`, `azure`, `onprem`, `alibaba`)
- `CYMBAL_BRANDING` - Use Cymbal Shops branding (default: `false`)
- `ENABLE_ASSISTANT` - Enable AI assistant feature (default: `false`)
- `ENABLE_SINGLE_SHARED_SESSION` - Use single shared session for all users (default: `false`, for demos only)

### ProductCatalogService-Specific

- `EXTRA_LATENCY` - Add artificial latency to requests (for testing)

### CartDatabase Configuration

- `CART_SERVICE_ADDR` - Cart service address for frontend/checkout

## Security Best Practices

### For Development/Demo
```bash
# Minimal secure configuration
ENABLE_RATE_LIMITING=true
RATE_LIMIT_AI=20
RATE_LIMIT_POST=100
RATE_LIMIT_GET=200
ENABLE_GRPC_TLS=false
```

### For Production
```bash
# Production-grade security
ENABLE_RATE_LIMITING=true
RATE_LIMIT_AI=10
RATE_LIMIT_POST=60
RATE_LIMIT_GET=120
ENABLE_GRPC_TLS=true
ENABLE_HTTPS=true
ALLOYDB_USER=cartservice_user  # Dedicated user, not postgres
ALLOYDB_READ_IP=<read-pool-ip>  # Enable read/write separation
```

### gRPC TLS Configuration

**Development/Testing** (no certificates):
```bash
ENABLE_GRPC_TLS=false
```

**Testing with self-signed certificates**:
```bash
ENABLE_GRPC_TLS=skip-verify  # WARNING: Only for testing!
```

**Production**:
```bash
ENABLE_GRPC_TLS=system  # Use system CA certificates
# OR
ENABLE_GRPC_TLS=custom
GRPC_TLS_CA_CERT=/path/to/ca-certificates.crt
```

## Helm Values Mapping

These environment variables are controlled by Helm values.yaml:

```yaml
securityFeatures:
  rateLimiting:
    enabled: true              → ENABLE_RATE_LIMITING=true
    aiLimit: 20                → RATE_LIMIT_AI=20
    postLimit: 100             → RATE_LIMIT_POST=100
    getLimit: 200              → RATE_LIMIT_GET=200
  grpcTls:
    enabled: false             → (no ENABLE_GRPC_TLS set)
    mode: "system"             → ENABLE_GRPC_TLS=system
```

## Validation

All environment variables are validated at service startup:

- ✅ **PORT**: Must be 1-65535
- ✅ **Required variables**: Services exit with error if missing required vars
- ✅ **Rate limits**: Must be positive integers
- ✅ **TLS mode**: Must be valid value (true/system/skip-verify/custom/false)

Check logs on startup for validation errors:
```bash
kubectl logs -l app=frontend | grep -i "error\|invalid"
```

## Monitoring

Security-related log messages to monitor:

- `Rate limit exceeded` - Rate limiting violations
- `Invalid prompt rejected` - AI input validation failures
- `Invalid image URL rejected` - Image URL validation failures
- `CSRF token validation failed` - CSRF attack attempts
- `Using TLS` / `Using insecure` - gRPC connection security mode

Example monitoring query:
```bash
# View rate limit violations
kubectl logs -l app=frontend | grep "Rate limit exceeded"

# View security validation failures
kubectl logs -l app=shoppingassistantservice | grep "Invalid.*rejected"
```

## Quick Reference

| Feature | Environment Variable | Default | Demo Value | Production Value |
|---------|---------------------|---------|------------|------------------|
| Rate Limiting | `ENABLE_RATE_LIMITING` | `false` | `true` | `true` |
| AI Rate Limit | `RATE_LIMIT_AI` | `10` | `20` | `10` |
| POST Rate Limit | `RATE_LIMIT_POST` | `60` | `100` | `60` |
| GET Rate Limit | `RATE_LIMIT_GET` | `120` | `200` | `120` |
| gRPC TLS | `ENABLE_GRPC_TLS` | `false` | `false` | `system` |
| HTTPS/HSTS | `ENABLE_HTTPS` | `false` | `false` | `true` |
| DB User | `ALLOYDB_USER` | `postgres` | `postgres` | `dedicated_user` |

## Troubleshooting

**Rate limiting too strict?**
```bash
helm upgrade myboutique ./helm-chart \
  --set securityFeatures.rateLimiting.aiLimit=100 \
  --set securityFeatures.rateLimiting.postLimit=500
```

**gRPC TLS connection failures?**
```bash
# Check if services have TLS enabled
kubectl logs -l app=frontend | grep "gRPC"

# Temporarily disable for debugging
helm upgrade myboutique ./helm-chart \
  --set securityFeatures.grpcTls.enabled=false
```

**CSRF validation failures?**
- Ensure cookies are enabled in browser
- Check for proxy/CDN stripping cookies
- Verify SameSite cookie support in browser

For more information, see [Security Features Documentation](./helm-chart/SECURITY_FEATURES.md).
