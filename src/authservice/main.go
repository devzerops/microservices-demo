package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/GoogleCloudPlatform/microservices-demo/src/authservice/genproto"
	"github.com/GoogleCloudPlatform/microservices-demo/src/authservice/storage"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	log        *logrus.Logger
	jwtSecret  []byte
	userStore  *storage.UserStore
)

const (
	defaultJWTSecret = "your-secret-key-change-in-production" // Should be set via env var
	tokenExpiration  = 24 * time.Hour // 24 hours
)

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

type authService struct {
	pb.UnimplementedAuthServiceServer
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// generateToken generates a JWT token for a user
func generateToken(user *pb.User) (string, error) {
	claims := &Claims{
		UserID: user.UserId,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "authservice",
			Subject:   user.UserId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// validateToken validates a JWT token and returns the claims
func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.WithFields(logrus.Fields{
		"email": req.Email,
		"name":  req.Name,
	}).Info("User registration request")

	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Create user
	user, err := userStore.CreateUser(req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		log.WithError(err).Error("Failed to create user")
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		log.WithError(err).Error("Failed to generate token")
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	log.WithField("user_id", user.UserId).Info("User registered successfully")

	return &pb.RegisterResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.WithField("email", req.Email).Info("User login request")

	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Authenticate user
	user, err := userStore.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) || errors.Is(err, storage.ErrInvalidPassword) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		log.WithError(err).Error("Failed to authenticate user")
		return nil, status.Error(codes.Internal, "failed to authenticate user")
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		log.WithError(err).Error("Failed to generate token")
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	log.WithField("user_id", user.UserId).Info("User logged in successfully")

	return &pb.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *authService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	claims, err := validateToken(req.Token)
	if err != nil {
		log.WithError(err).Debug("Invalid token")
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	// Get user information
	user, err := userStore.GetUser(claims.UserID)
	if err != nil {
		log.WithError(err).Warn("User not found for valid token")
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		User:  user,
	}, nil
}

// GetUser retrieves user information by user ID
func (s *authService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := userStore.GetUser(req.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		log.WithError(err).Error("Failed to get user")
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return user, nil
}

func initTracing() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func main() {
	port := "8081"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	// Get JWT secret from environment or use default
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = defaultJWTSecret
		log.Warn("Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}
	jwtSecret = []byte(secret)

	// Initialize user store
	userStore = storage.NewUserStore()
	log.Info("User store initialized")

	// Initialize tracing if OTEL_SERVICE_NAME is set
	if os.Getenv("OTEL_SERVICE_NAME") != "" {
		tp, err := initTracing()
		if err != nil {
			log.WithError(err).Warn("Failed to initialize tracing")
		} else {
			defer func() {
				if err := tp.Shutdown(context.Background()); err != nil {
					log.WithError(err).Error("Error shutting down tracer provider")
				}
			}()
			log.Info("Tracing initialized")
		}
	}

	// Create gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var srv *grpc.Server
	if os.Getenv("DISABLE_TRACING") == "" {
		srv = grpc.NewServer(
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		)
	} else {
		srv = grpc.NewServer()
	}

	pb.RegisterAuthServiceServer(srv, &authService{})

	log.Infof("AuthService listening on port %s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
