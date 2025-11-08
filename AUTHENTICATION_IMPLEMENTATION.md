# Authentication Implementation Summary

## Overview
This document describes the implementation of Phase 2 (Frontend Integration) and Phase 3 (Service Integration) for user authentication in the microservices-demo application.

## Phase 2: Frontend Integration ✅

### 1. AuthService (New Microservice)
**Location:** `src/authservice/`

**Features:**
- User registration with email and password
- User login with JWT token generation
- Token validation
- In-memory user storage (can be extended to database)
- Password hashing with bcrypt
- JWT token signing with configurable secret

**Technologies:**
- Go 1.23
- JWT (golang-jwt/jwt/v5)
- bcrypt for password hashing
- gRPC for inter-service communication

**Key Files:**
- `main.go` - Main service implementation with gRPC handlers
- `storage/storage.go` - In-memory user store
- `Dockerfile` - Container image definition
- `go.mod` - Go module dependencies

### 2. Frontend Updates

**Login/Signup Pages:**
- `src/frontend/templates/login.html` - Login page
- `src/frontend/templates/signup.html` - Registration page
- `src/frontend/templates/header.html` - Updated to show user info and login/logout buttons

**Handlers:**
- `src/frontend/auth_handlers.go` - New file with authentication handlers
  - `loginHandler` - Handles GET (show form) and POST (process login)
  - `signupHandler` - Handles GET (show form) and POST (process registration)
  - `getCurrentUser` - Helper to retrieve authenticated user from JWT token

**Middleware:**
- `src/frontend/middleware.go` - Added `ctxKeyUser` for user context
- `src/frontend/handlers.go` - Updated `injectCommonTemplateData` to include user info in all templates

**Routes:**
- `GET/POST /login` - Login page and handler
- `GET/POST /signup` - Signup page and handler
- `GET /logout` - Logout (clears auth token cookie)

### 3. JWT Token Management

**Cookie Configuration:**
- Name: `shop_auth-token`
- Max Age: 48 hours
- HttpOnly: true (prevents XSS attacks)
- SameSite: Lax (CSRF protection)
- Path: / (site-wide)

**Token Claims:**
- User ID
- Email
- Issued At (iat)
- Expires At (exp)
- Not Before (nbf)
- Subject (sub)
- Issuer

### 4. Anonymous vs Authenticated Users

**Anonymous Users:**
- Can browse products
- Can add items to cart (session-based)
- Cannot checkout without login

**Authenticated Users:**
- See personalized greeting in header
- User-specific cart (linked to user_id)
- Can complete checkout
- Order history (future enhancement)

## Phase 3: Service Integration (Foundation)

### Architecture Updates

**Proto Definitions:**
Updated `protos/demo.proto` with new AuthService:
```protobuf
service AuthService {
    rpc Register(RegisterRequest) returns (RegisterResponse) {}
    rpc Login(LoginRequest) returns (LoginResponse) {}
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse) {}
    rpc GetUser(GetUserRequest) returns (User) {}
}
```

### Service Connections

**Frontend Service:**
- Added `authSvcAddr` and `authSvcConn` to connect to AuthService
- Environment variable: `AUTH_SERVICE_ADDR=authservice:8081`

**Future Integration Points:**
The following services are now ready for user-context integration:

1. **CartService** (C# .NET)
   - Already uses `user_id` parameter
   - Currently treats all user_ids as trusted
   - Can be enhanced to validate JWT tokens via gRPC interceptors

2. **CheckoutService** (Go)
   - Already uses `user_id` from PlaceOrderRequest
   - Can be enhanced to:
     - Store order history per user
     - Retrieve user's past orders
     - User-specific checkout preferences

3. **RecommendationService** (Python)
   - Already uses `user_id` parameter
   - Can be enhanced to:
     - Provide personalized recommendations
     - Track user purchase history
     - ML-based product suggestions

## Kubernetes Deployment

### New Resources

**AuthService Deployment:**
- File: `kubernetes-manifests/authservice.yaml`
- Service port: 8081 (gRPC)
- Resource limits: 200m CPU, 128Mi memory
- Security: Non-root user, read-only root filesystem
- JWT Secret: Kubernetes Secret (`jwt-secret`)

**Frontend Updates:**
- Added `AUTH_SERVICE_ADDR` environment variable
- Points to `authservice:8081`

**Kustomization:**
- Updated `kubernetes-manifests/kustomization.yaml` to include `authservice.yaml`

## Security Features

### Implemented:
1. **Password Security:**
   - bcrypt hashing with default cost factor
   - Passwords never stored in plaintext

2. **JWT Security:**
   - Signed tokens with HS256
   - Configurable secret via environment variable
   - Token expiration (24 hours)
   - Token validation on each request

3. **Cookie Security:**
   - HttpOnly flag (prevents XSS)
   - SameSite=Lax (prevents CSRF)
   - Secure flag ready for HTTPS

4. **Container Security:**
   - Non-root user execution
   - Read-only root filesystem
   - Dropped all capabilities
   - No privilege escalation

### Security Recommendations for Production:

1. **JWT Secret:**
   - Generate a strong random key (256+ bits)
   - Store in Kubernetes Secret or secret management service
   - Rotate regularly

2. **HTTPS:**
   - Enable TLS/HTTPS for frontend
   - Set Secure flag on cookies
   - Use cert-manager for certificate management

3. **Database:**
   - Replace in-memory storage with persistent database
   - Use connection pooling
   - Enable encryption at rest

4. **Rate Limiting:**
   - Add rate limiting to login/signup endpoints
   - Implement account lockout after failed attempts

5. **Token Refresh:**
   - Implement refresh token mechanism
   - Shorter access token expiration (15 minutes)
   - Longer refresh token expiration (7 days)

## Testing

### Manual Testing Steps:

1. **User Registration:**
   ```bash
   curl -X POST http://localhost:8080/signup \
     -d "name=John Doe" \
     -d "email=john@example.com" \
     -d "password=secret123" \
     -d "confirm_password=secret123"
   ```

2. **User Login:**
   ```bash
   curl -X POST http://localhost:8080/login \
     -d "email=john@example.com" \
     -d "password=secret123" \
     -c cookies.txt
   ```

3. **Access Protected Resources:**
   ```bash
   curl http://localhost:8080/ -b cookies.txt
   ```

4. **Logout:**
   ```bash
   curl http://localhost:8080/logout -b cookies.txt
   ```

### UI Testing:
1. Navigate to http://localhost:8080/
2. Click "Sign Up" and create an account
3. Verify greeting appears in header
4. Browse products and add to cart
5. Click "Logout" and verify session cleared

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (Go)                        │
│                  HTTP + gRPC Client                     │
│                                                         │
│  Routes:                                                │
│  - GET/POST /login     → AuthService.Login            │
│  - GET/POST /signup    → AuthService.Register         │
│  - GET /logout         → Clear JWT cookie              │
│  - All pages           → Inject user context          │
└─────────────┬──────────────────────────────────────────┘
              │
              │ gRPC (port 8081)
              │
         ┌────▼─────────────────────────────────────┐
         │       AuthService (Go)                   │
         │                                           │
         │  - User Registration                      │
         │  - User Login (JWT generation)           │
         │  - Token Validation                       │
         │  - User Lookup                            │
         │                                           │
         │  Storage: In-Memory (UserStore)          │
         │  - Users map[user_id]UserData            │
         │  - Emails map[email]user_id              │
         └───────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│              Ready for Integration                      │
├─────────────────────────────────────────────────────────┤
│  CartService       - User-specific carts               │
│  CheckoutService   - User order history                │
│  RecommendationSvc - Personalized recommendations      │
└─────────────────────────────────────────────────────────┘
```

## Next Steps (Future Enhancements)

### Phase 3 Completion:
1. **CartService Integration:**
   - Add JWT validation interceptor
   - Ensure cart isolation per authenticated user
   - Migrate anonymous cart to user cart on login

2. **CheckoutService Integration:**
   - Store order history with user_id
   - Add endpoint to retrieve user's past orders
   - User-specific shipping addresses

3. **Database Integration:**
   - Replace in-memory storage with PostgreSQL/MySQL
   - User table schema
   - Migration scripts
   - Connection pooling

4. **Advanced Features:**
   - Email verification
   - Password reset flow
   - Social login (OAuth)
   - Two-factor authentication (2FA)
   - User profile management
   - Admin user roles

### Performance & Scalability:
1. Token caching with Redis
2. Database read replicas
3. Horizontal scaling of AuthService
4. Session management with distributed cache

## Files Changed/Created

### New Files:
- `src/authservice/main.go`
- `src/authservice/storage/storage.go`
- `src/authservice/go.mod`
- `src/authservice/Dockerfile`
- `src/authservice/.dockerignore`
- `src/authservice/genproto.sh`
- `src/authservice/README.md`
- `src/frontend/auth_handlers.go`
- `src/frontend/templates/login.html`
- `src/frontend/templates/signup.html`
- `kubernetes-manifests/authservice.yaml`

### Modified Files:
- `protos/demo.proto` (Added AuthService definitions)
- `src/frontend/main.go` (Added auth service connection and routes)
- `src/frontend/handlers.go` (Updated injectCommonTemplateData)
- `src/frontend/middleware.go` (Added ctxKeyUser)
- `src/frontend/templates/header.html` (Added login/logout UI)
- `kubernetes-manifests/frontend.yaml` (Added AUTH_SERVICE_ADDR)
- `kubernetes-manifests/kustomization.yaml` (Added authservice)

## Conclusion

This implementation provides a solid foundation for user authentication in the microservices-demo application. The system now supports:

✅ User registration and login
✅ JWT-based authentication
✅ Secure password storage
✅ Anonymous and authenticated user sessions
✅ Frontend integration with login/signup UI
✅ Ready for backend service integration
✅ Production-ready Kubernetes deployment
✅ Security best practices

The architecture is designed to be scalable and can be enhanced with additional features as needed.
