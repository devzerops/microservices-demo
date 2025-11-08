# Auth Service

Authentication service for user registration, login, and JWT token management.

## Features

- User registration with email and password
- User login with JWT token generation
- Token validation
- In-memory user storage (can be extended to use database)

## Environment Variables

- `PORT` - gRPC server port (default: 8081)
- `JWT_SECRET` - Secret key for JWT token signing (required in production)
- `OTEL_SERVICE_NAME` - Enable OpenTelemetry tracing
- `DISABLE_TRACING` - Disable distributed tracing
