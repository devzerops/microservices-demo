# Security Guide - Microservices Demo

This document outlines the security measures implemented in this project and provides recommendations for production deployment.

## Table of Contents

- [Recent Security Fixes](#recent-security-fixes)
- [Remaining Security Considerations](#remaining-security-considerations)
- [Security Best Practices](#security-best-practices)
- [Reporting Security Issues](#reporting-security-issues)
- [Security Testing](#security-testing)

---

## Recent Security Fixes

The following security vulnerabilities have been identified and fixed in commit `844a64f`:

### Critical Severity

#### 1. SQL Injection (CWE-89) - FIXED ✅

**Location**: `src/productcatalogservice/catalog_loader.go:132`

**Vulnerability**: Table name from environment variable was concatenated directly into SQL query without validation.

```go
// BEFORE (Vulnerable)
query := "SELECT ... FROM " + pgTableName
```

**Fix Applied**:
- Added input validation with regex pattern matching
- Implemented `pgx.Identifier.Sanitize()` for safe SQL identifier handling
- Validates table name format: `^[a-zA-Z_][a-zA-Z0-9_]*$`
- Maximum length check (63 characters)

```go
// AFTER (Secure)
if err := validateTableName(pgTableName); err != nil {
    return err
}
query := fmt.Sprintf("SELECT ... FROM %s", pgx.Identifier{pgTableName}.Sanitize())
```

**Impact**: Prevents malicious SQL injection through `ALLOYDB_TABLE_NAME` environment variable.

---

### High Severity

#### 2. Server-Side Request Forgery (CWE-918) - FIXED ✅

**Location**: `src/frontend/packaging_info.go:52-54`

**Vulnerability**: Product ID was used to construct URLs without validation, allowing SSRF attacks.

```go
// BEFORE (Vulnerable)
url := packagingServiceUrl + "/" + productId
resp, err := http.Get(url)
```

**Fix Applied**:
- Input validation for product IDs (alphanumeric + hyphens only)
- URL construction using `url.JoinPath()`
- Host verification to prevent URL manipulation
- HTTP client timeout (10 seconds)

```go
// AFTER (Secure)
if err := validateProductId(productId); err != nil {
    return nil, err
}
fullURL := baseURL.JoinPath(productId).String()
// Host verification
if finalURL.Host != baseURL.Host {
    return nil, fmt.Errorf("URL host mismatch: potential SSRF attack")
}
```

**Impact**: Prevents SSRF attacks, internal port scanning, and access to metadata endpoints.

---

#### 3. Missing Input Validation (CWE-20) - FIXED ✅

**Location**: `src/shoppingassistantservice/shoppingassistantservice.py:68, 79`

**Vulnerability**: Direct access to JSON fields without validation.

```python
# BEFORE (Vulnerable)
prompt = request.json['message']
image_url = request.json['image']
```

**Fix Applied**:
- Content-Type validation (must be `application/json`)
- Required field validation (`message`, `image`)
- Type checking (must be non-empty strings)
- Proper error responses (HTTP 400)

```python
# AFTER (Secure)
if not request.is_json:
    return jsonify({'error': 'Content-Type must be application/json'}), 400
if 'message' not in request.json:
    return jsonify({'error': 'Missing required field: message'}), 400
if not prompt or not isinstance(prompt, str):
    return jsonify({'error': 'message must be a non-empty string'}), 400
```

**Impact**: Prevents KeyError crashes, type confusion attacks, and improves API robustness.

---

#### 4. Undefined Variable / Runtime Crash - FIXED ✅

**Location**: `src/frontend/handlers.go:406`

**Vulnerability**: Used undefined `log` variable causing runtime panic.

**Fix Applied**:
```go
// Added at the beginning of assistantHandler
log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
```

**Impact**: Prevents service crashes when accessing the assistant page.

---

#### 5. Context Propagation Failure (CWE-705) - FIXED ✅

**Location**: `src/checkoutservice/main.go:361`

**Vulnerability**: Using `context.TODO()` instead of propagating actual context.

```go
// BEFORE
result, err := currencyClient.Convert(context.TODO(), ...)
```

**Fix Applied**:
```go
// AFTER
result, err := currencyClient.Convert(ctx, ...)
```

**Impact**: Enables proper timeout, cancellation, and distributed tracing context propagation.

---

### Medium Severity

#### 6. Missing Error Handling - FIXED ✅

**Location**: `src/frontend/handlers.go:213, 327, 332-334`

**Vulnerability**: Ignored parse errors leading to zero values being used silently.

```go
// BEFORE
quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
zipCode, _ := strconv.ParseInt(r.FormValue("zip_code"), 10, 32)
```

**Fix Applied**:
```go
// AFTER
quantity, err := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
if err != nil {
    renderHTTPError(log, r, w, errors.Wrap(err, "invalid quantity format"), http.StatusBadRequest)
    return
}
```

**Impact**: Prevents processing invalid inputs, improves user feedback, prevents business logic errors.

---

#### 7. Resource Exhaustion (CWE-400) - FIXED ✅

**Location**: `src/frontend/handlers.go:472`, `src/frontend/packaging_info.go:54`

**Vulnerability**: HTTP clients without timeouts vulnerable to slowloris attacks.

**Fix Applied**:
```go
var httpClientWithTimeout = &http.Client{
    Timeout: 30 * time.Second,  // frontend handlers
}

var packagingHTTPClient = &http.Client{
    Timeout: 10 * time.Second,  // packaging service
}
```

**Impact**: Prevents resource exhaustion, ensures bounded waiting times, protects against slow HTTP attacks.

---

#### 8. Resource Leak - FIXED ✅

**Location**: `src/frontend/handlers.go:498`

**Vulnerability**: HTTP response body not closed, causing memory leaks.

**Fix Applied**:
```go
res, err := httpClientWithTimeout.Do(req)
if err != nil {
    return
}
defer res.Body.Close()  // Added
```

**Impact**: Prevents memory leaks, ensures proper connection pool management.

---

#### 9. Weak Random Number Generation (CWE-338) - FIXED ✅

**Locations**:
- `src/shippingservice/tracker.go:19, 29, 45`
- `src/frontend/handlers.go:573`
- `src/adservice/src/main/java/hipstershop/AdService.java:141`

**Vulnerability**: Using `math/rand`, `random.Random`, `java.util.Random` for generating tracking IDs and selecting items.

**Fix Applied**:

**Go (shippingservice, frontend)**:
```go
// BEFORE
import "math/rand"
rand.Seed(time.Now().UnixNano())
n := rand.Intn(max)

// AFTER
import "crypto/rand"
import "math/big"
n, err := rand.Int(rand.Reader, big.NewInt(max))
```

**Java (adservice)**:
```java
// BEFORE
import java.util.Random;
private static final Random random = new Random();

// AFTER
import java.security.SecureRandom;
private static final SecureRandom random = new SecureRandom();
```

**Impact**: Prevents predictable random numbers, improves tracking ID security, prevents pattern-based attacks.

---

## Remaining Security Considerations

The following security considerations require attention for production deployment:

### 🔴 High Priority

#### 1. Insecure gRPC Connections

**Status**: ⚠️ Not Fixed (Demo Application)

**Location**: All microservices

**Issue**: All gRPC connections use insecure channels (no TLS):

```go
// Go services
grpc.Dial(addr, grpc.WithInsecure())

// Node.js services
grpc.credentials.createInsecure()

// Python services
grpc.insecure_channel(addr)
```

**Risk**:
- All data transmitted in plaintext
- No authentication between services
- Vulnerable to MITM attacks
- Credit card data exposed on network

**Recommendation for Production**:

1. **Implement mTLS (Mutual TLS)**:
```go
// Example for Go
creds, err := credentials.NewClientTLSFromFile("ca.crt", "")
conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
```

2. **Use service mesh** (e.g., Istio):
   - Automatic mTLS between services
   - Certificate management
   - Traffic encryption

3. **Certificate Management**:
   - Use cert-manager for Kubernetes
   - Rotate certificates regularly
   - Use short-lived certificates

**References**:
- [gRPC Authentication Guide](https://grpc.io/docs/guides/auth/)
- [Istio Security](https://istio.io/latest/docs/concepts/security/)

---

#### 2. Database Security (AlloyDB)

**Status**: ⚠️ Requires Review

**Location**: `src/cartservice/src/cartstore/AlloyDBCartStore.cs:42-46`

**Issues**:

```csharp
// TODO: Create a separate user for connecting within the application
// rather than using our superuser
string alloyDBUser = "postgres";

// TODO: Consider splitting workloads into read vs. write and take
// advantage of the AlloyDB read pools
```

**Recommendations**:

1. **Principle of Least Privilege**:
   - Create dedicated database users per service
   - Grant minimum required permissions
   - Example permissions:
     ```sql
     -- Cart service user
     CREATE USER cartservice_user WITH PASSWORD 'strong_password';
     GRANT SELECT, INSERT, UPDATE, DELETE ON cart_table TO cartservice_user;
     -- No DROP, CREATE, ALTER permissions
     ```

2. **Connection Pooling**:
   - Implement read/write splitting
   - Use AlloyDB read pools for read-only queries
   - Configure connection limits

3. **Password Management**:
   - ✅ Currently using Google Secret Manager (good!)
   - Rotate passwords regularly
   - Use strong password requirements

---

#### 3. Secret Management

**Status**: ⚠️ Partial Implementation

**Current State**:
- ✅ Using Google Secret Manager for database passwords
- ⚠️ API keys may be in environment variables
- ⚠️ No secret rotation policy

**Recommendations**:

1. **Centralize All Secrets**:
```go
// Example: Load API keys from Secret Manager
apiKey, err := getSecretPayload(projectID, "api-key-secret", "latest")
```

2. **Implement Secret Rotation**:
   - Set up automatic rotation for database passwords
   - Implement zero-downtime rotation
   - Use Secret Manager versioning

3. **Never Commit Secrets**:
   - ✅ `.env` files in `.gitignore`
   - Use git-secrets or similar tools
   - Scan for accidentally committed secrets

---

### 🟡 Medium Priority

#### 4. Rate Limiting

**Status**: ❌ Not Implemented

**Risk**: Services vulnerable to DoS attacks

**Recommendations**:

1. **API Gateway Level**:
```yaml
# Example: Kong/Nginx rate limiting
apiVersion: configuration.konghq.com/v1
kind: KongPlugin
metadata:
  name: rate-limit
config:
  minute: 60
  policy: local
```

2. **Application Level** (Python example):
```python
from flask_limiter import Limiter

limiter = Limiter(
    app,
    key_func=lambda: request.remote_addr,
    default_limits=["100 per minute"]
)

@app.route("/", methods=['POST'])
@limiter.limit("10 per minute")
def talkToGemini():
    # ...
```

3. **Service Mesh** (Istio example):
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: frontend-rate-limit
spec:
  host: frontend
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 100
        maxRequestsPerConnection: 10
```

---

#### 5. Input Sanitization for Logging

**Status**: ⚠️ Needs Review

**Risk**: Log injection attacks

**Example Vulnerable Code**:
```go
log.Printf("User input: %s", userInput) // Could contain newlines, control chars
```

**Recommendations**:

1. **Sanitize Before Logging**:
```go
import "strings"

func sanitizeForLog(input string) string {
    // Remove control characters and newlines
    input = strings.Map(func(r rune) rune {
        if r < 32 || r == 127 {
            return -1
        }
        return r
    }, input)
    return input
}

log.Printf("User input: %s", sanitizeForLog(userInput))
```

2. **Use Structured Logging**:
```go
log.WithFields(logrus.Fields{
    "user_id": userID,
    "action": "purchase",
}).Info("Order placed")
// Fields are automatically escaped
```

---

#### 6. Content Security Policy (CSP)

**Status**: ❌ Not Implemented

**Location**: Frontend service

**Recommendation**:

```go
// Add CSP headers to all responses
w.Header().Set("Content-Security-Policy",
    "default-src 'self'; "+
    "script-src 'self' 'unsafe-inline'; "+
    "style-src 'self' 'unsafe-inline'; "+
    "img-src 'self' data: https:; "+
    "font-src 'self' data:; "+
    "connect-src 'self'")
```

---

### 🟢 Low Priority

#### 7. Security Headers

**Status**: ⚠️ Partial Implementation

**Recommendations**:

```go
func setSecurityHeaders(w http.ResponseWriter) {
    // Prevent clickjacking
    w.Header().Set("X-Frame-Options", "DENY")

    // Enable browser XSS protection
    w.Header().Set("X-XSS-Protection", "1; mode=block")

    // Prevent MIME sniffing
    w.Header().Set("X-Content-Type-Options", "nosniff")

    // Enforce HTTPS
    w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

    // Referrer policy
    w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

    // Permissions policy
    w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}
```

---

#### 8. Dependency Vulnerabilities

**Status**: ⚠️ Requires Regular Scanning

**Recommendations**:

1. **Automated Scanning**:
```bash
# Go
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Java
./gradlew dependencyCheckAnalyze

# Python
pip install safety
safety check

# Node.js
npm audit
```

2. **Container Scanning**:
```bash
# Using Trivy
trivy image gcr.io/google-samples/microservices-demo/frontend:latest
```

3. **Continuous Monitoring**:
   - Enable GitHub Dependabot
   - Use Snyk or similar tools
   - Set up automated PR creation for updates

---

## Security Best Practices

### Development

1. **Code Review**:
   - All changes require security review
   - Use security-focused linters
   - Follow OWASP guidelines

2. **Static Analysis**:
```bash
# Go
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Python
pip install bandit
bandit -r src/

# Java
./gradlew spotbugsMain
```

3. **Secrets in Code**:
   - Never commit secrets
   - Use environment variables or secret managers
   - Scan commits with git-secrets

### Deployment

1. **Network Segmentation**:
   - Use network policies in Kubernetes
   - Isolate services by namespace
   - Restrict egress traffic

2. **Pod Security**:
```yaml
# Example Pod Security Policy
apiVersion: v1
kind: Pod
metadata:
  name: frontend
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
  - name: frontend
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
```

3. **Image Security**:
   - Use minimal base images (distroless, alpine)
   - Sign container images
   - Scan for vulnerabilities before deployment

### Monitoring

1. **Security Monitoring**:
   - Enable audit logging
   - Monitor for unusual patterns
   - Set up alerts for security events

2. **Incident Response**:
   - Define incident response plan
   - Regular security drills
   - Maintain security runbooks

---

## Reporting Security Issues

If you discover a security vulnerability in this project:

1. **DO NOT** open a public GitHub issue
2. **DO NOT** disclose the vulnerability publicly
3. **DO** report it privately to the maintainers

### Reporting Process

Email: [security@example.com] (replace with actual contact)

Include in your report:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will:
- Acknowledge receipt within 48 hours
- Provide a detailed response within 7 days
- Keep you informed of our progress

### Security Response Timeline

- **Critical**: Fix within 24 hours
- **High**: Fix within 1 week
- **Medium**: Fix within 1 month
- **Low**: Fix in next scheduled release

---

## Security Testing

### Automated Security Testing

1. **SAST (Static Application Security Testing)**:
```bash
# Run security scanners
make security-scan

# Or individually:
gosec ./...                    # Go
bandit -r src/                 # Python
npm audit                      # Node.js
./gradlew dependencyCheck      # Java
```

2. **DAST (Dynamic Application Security Testing)**:
```bash
# Example using OWASP ZAP
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t http://frontend:8080
```

3. **Container Scanning**:
```bash
# Scan all service images
make container-scan

# Or individually:
trivy image gcr.io/.../frontend:latest
```

### Manual Security Testing

1. **SQL Injection Testing**:
```bash
# Test table name validation
export ALLOYDB_TABLE_NAME="products'; DROP TABLE users; --"
# Should fail with validation error
```

2. **SSRF Testing**:
```bash
# Test product ID validation
curl -X GET "http://frontend:8080/product/../../etc/passwd"
# Should return 400 Bad Request

curl -X GET "http://frontend:8080/product/@169.254.169.254"
# Should return 400 Bad Request
```

3. **Input Validation Testing**:
```bash
# Test shopping assistant validation
curl -X POST http://shopping-assistant:8080/ \
  -H "Content-Type: text/plain" \
  -d "invalid"
# Should return 400 Bad Request

curl -X POST http://shopping-assistant:8080/ \
  -H "Content-Type: application/json" \
  -d '{}'
# Should return 400 Bad Request: Missing required field
```

### Penetration Testing

For production deployments, conduct regular penetration testing:

1. **Third-party Pentest**: Annually
2. **Internal Security Review**: Quarterly
3. **Automated Scanning**: Weekly

---

## Security Compliance

### OWASP Top 10 Coverage

| Risk | Status | Notes |
|------|--------|-------|
| A01:2021 Broken Access Control | ⚠️ Partial | No authentication between services (mTLS needed) |
| A02:2021 Cryptographic Failures | ✅ Fixed | Using SecureRandom, crypto/rand |
| A03:2021 Injection | ✅ Fixed | SQL injection, SSRF fixed |
| A04:2021 Insecure Design | ⚠️ Review | Architecture review recommended |
| A05:2021 Security Misconfiguration | ⚠️ Partial | Insecure gRPC, need security headers |
| A06:2021 Vulnerable Components | ⚠️ Ongoing | Regular dependency updates needed |
| A07:2021 Authentication Failures | ⚠️ N/A | Demo app - no user authentication |
| A08:2021 Data Integrity Failures | ⚠️ Partial | Need request signing |
| A09:2021 Logging Failures | ✅ Good | Structured logging implemented |
| A10:2021 Server-Side Request Forgery | ✅ Fixed | SSRF vulnerability fixed |

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-01-XX | Initial security documentation |
| 1.1 | 2025-01-XX | Added security fixes from commit 844a64f |

---

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE/SANS Top 25](https://cwe.mitre.org/top25/)
- [Google Cloud Security Best Practices](https://cloud.google.com/security/best-practices)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [gRPC Security Documentation](https://grpc.io/docs/guides/auth/)

---

**Note**: This is a demo application. For production use, implement all recommendations in the "Remaining Security Considerations" section and conduct a thorough security audit.
