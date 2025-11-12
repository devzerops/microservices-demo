# Security Features Configuration

This Helm chart includes enhanced security features that can be easily enabled for demo/practice purposes.

## Quick Start - Deploy with Security Features

```bash
# Deploy with all security features enabled (recommended for demo)
helm install myboutique ./helm-chart

# The following security features are automatically enabled:
# ✅ Rate Limiting (relaxed limits for demo)
# ✅ CSRF Protection (always enabled)
# ✅ SQL Injection Prevention (always enabled)
# ✅ Input Validation (always enabled)
# ✅ Graceful Shutdown (always enabled)
```

## Security Features Included

### 1. **Rate Limiting** (Enabled by default)
Protects against API abuse and DoS attacks with per-session limits.

**Default Settings (relaxed for demo):**
- AI Assistant endpoint (`/bot`): 20 requests/minute
- POST endpoints: 100 requests/minute
- GET endpoints: 200 requests/minute

**Configuration:**
```yaml
securityFeatures:
  rateLimiting:
    enabled: true    # Set to false to disable
    aiLimit: 20      # Adjust as needed
    postLimit: 100   # Adjust as needed
    getLimit: 200    # Adjust as needed
```

**To disable:**
```bash
helm install myboutique ./helm-chart --set securityFeatures.rateLimiting.enabled=false
```

### 2. **CSRF Protection** (Always Enabled)
Protects POST endpoints from Cross-Site Request Forgery attacks.
- Automatically generates secure tokens
- Validates all form submissions
- No configuration needed

### 3. **SQL Injection Prevention** (Always Enabled)
All database queries use parameterized statements.
- No configuration needed
- Works automatically

### 4. **gRPC TLS Encryption** (Disabled by default)
Encrypts inter-service gRPC communication.

**Note:** Requires TLS certificates to be configured. Disabled by default for easy demo setup.

**To enable (requires certificates):**
```yaml
securityFeatures:
  grpcTls:
    enabled: true
    mode: "system"  # Use system CA certificates
```

**For testing without certificates:**
```yaml
securityFeatures:
  grpcTls:
    enabled: true
    mode: "skip-verify"  # WARNING: Testing only!
```

### 5. **Other Security Features** (Always Enabled)
- Input validation on all user inputs
- Structured logging with security events
- Graceful shutdown with connection cleanup
- Resource leak prevention

## Advanced Configuration

### Custom Rate Limits

```bash
# Strict limits (production-like)
helm install myboutique ./helm-chart \
  --set securityFeatures.rateLimiting.aiLimit=10 \
  --set securityFeatures.rateLimiting.postLimit=60 \
  --set securityFeatures.rateLimiting.getLimit=120

# Very relaxed (for load testing)
helm install myboutique ./helm-chart \
  --set securityFeatures.rateLimiting.aiLimit=1000 \
  --set securityFeatures.rateLimiting.postLimit=5000 \
  --set securityFeatures.rateLimiting.getLimit=10000
```

### Enable gRPC TLS (Advanced)

```bash
# With system CA certificates (requires proper TLS setup)
helm install myboutique ./helm-chart \
  --set securityFeatures.grpcTls.enabled=true \
  --set securityFeatures.grpcTls.mode="system"

# For testing (skip certificate verification - NOT for production!)
helm install myboutique ./helm-chart \
  --set securityFeatures.grpcTls.enabled=true \
  --set securityFeatures.grpcTls.mode="skip-verify"
```

## Monitoring Security Features

Check if rate limiting is working:
```bash
# View logs for rate limit violations
kubectl logs -l app=frontend | grep "Rate limit exceeded"

# Check rate limit headers in responses
curl -I http://<frontend-url>/ | grep X-RateLimit
```

## Disabling Security Features

To deploy without enhanced security features (original behavior):
```bash
helm install myboutique ./helm-chart \
  --set securityFeatures.rateLimiting.enabled=false \
  --set securityFeatures.grpcTls.enabled=false
```

Note: CSRF protection, SQL injection prevention, and input validation remain active as they are implemented at the code level.

## Security Impact Summary

| Feature | Protection Against | Default Status | Config Required |
|---------|-------------------|----------------|-----------------|
| Rate Limiting | DoS, API abuse, brute force | ✅ Enabled | No |
| CSRF Protection | Cross-site request forgery | ✅ Enabled | No |
| SQL Injection Prevention | Database attacks | ✅ Enabled | No |
| Input Validation | Injection attacks, overflow | ✅ Enabled | No |
| gRPC TLS | Man-in-the-middle, sniffing | ❌ Disabled | Yes (certificates) |
| Graceful Shutdown | Resource leaks, data loss | ✅ Enabled | No |

## For Demo/Practice Use

The default configuration is optimized for easy demo deployment:
- ✅ **No certificates required** - gRPC TLS disabled by default
- ✅ **Relaxed rate limits** - Won't interfere with normal testing
- ✅ **All security features work** - CSRF, SQL protection, validation enabled
- ✅ **Easy to deploy** - Just `helm install` and it works!

Simply run:
```bash
helm install myboutique ./helm-chart
```

All security features except gRPC TLS will be active with demo-friendly settings.
