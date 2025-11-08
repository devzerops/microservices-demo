// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GoogleCloudPlatform/microservices-demo/src/authservice/auth"
	"github.com/GoogleCloudPlatform/microservices-demo/src/authservice/database"
	pb "github.com/GoogleCloudPlatform/microservices-demo/src/authservice/genproto"
)

// AuthServiceServer implements the AuthService gRPC server
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	db         *database.DB
	jwtManager *auth.JWTManager
	log        *logrus.Logger
}

// NewAuthServiceServer creates a new AuthServiceServer
func NewAuthServiceServer(db *database.DB, jwtManager *auth.JWTManager, log *logrus.Logger) *AuthServiceServer {
	return &AuthServiceServer{
		db:         db,
		jwtManager: jwtManager,
		log:        log,
	}
}

// SignUp handles user registration
func (s *AuthServiceServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	s.log.Infof("[SignUp] email=%s name=%s", req.Email, req.Name)

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, password, and name are required")
	}

	// Validate email format (basic check)
	if !strings.Contains(req.Email, "@") {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email format")
	}

	// Hash the password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.log.Errorf("failed to hash password: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to process password")
	}

	// Create user in database
	user, err := s.db.CreateUser(req.Email, passwordHash, req.Name)
	if err != nil {
		s.log.Errorf("failed to create user: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.log.Errorf("failed to generate token: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	s.log.Infof("[SignUp] successful for user %s", user.ID)

	return &pb.SignUpResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
		},
		Token: token,
	}, nil
}

// SignIn handles user authentication
func (s *AuthServiceServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	s.log.Infof("[SignIn] email=%s", req.Email)

	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	// Get user from database
	user, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		s.log.Warnf("user not found: %v", err)
		return nil, status.Errorf(codes.NotFound, "invalid email or password")
	}

	// Check password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		s.log.Warnf("password mismatch for user %s", user.ID)
		return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.log.Errorf("failed to generate token: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	s.log.Infof("[SignIn] successful for user %s", user.ID)

	return &pb.SignInResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
		},
		Token: token,
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	s.log.Debugf("[ValidateToken] validating token")

	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token is required")
	}

	// Validate the token
	claims, err := s.jwtManager.ValidateToken(req.Token)
	if err != nil {
		s.log.Warnf("invalid token: %v", err)
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// Get user from database
	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil {
		s.log.Warnf("user not found for token: %v", err)
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	s.log.Debugf("[ValidateToken] token valid for user %s", user.ID)

	return &pb.ValidateTokenResponse{
		Valid: true,
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
		},
	}, nil
}

// GetUser retrieves user information by ID
func (s *AuthServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.log.Infof("[GetUser] id=%s", req.Id)

	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	// Get user from database
	user, err := s.db.GetUserByID(req.Id)
	if err != nil {
		s.log.Warnf("user not found: %v", err)
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
		},
	}, nil
}
