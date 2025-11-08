# Auth Service

Authentication service for the microservices demo application.

## Features

- User registration (Sign up)
- User login (Sign in)
- JWT token generation and validation
- Password hashing with bcrypt
- PostgreSQL database for user storage

## API

The service provides both gRPC and HTTP endpoints for authentication operations.

### gRPC

- `SignUp`: Register a new user
- `SignIn`: Authenticate a user and receive a JWT token
- `ValidateToken`: Validate a JWT token
- `GetUser`: Get user information by ID

### Environment Variables

- `PORT`: gRPC server port (default: 8080)
- `HTTP_PORT`: HTTP server port (default: 8081)
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token generation
