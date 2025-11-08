package storage

import (
	"errors"
	"sync"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/authservice/genproto"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists      = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

// UserStore is an in-memory storage for users
type UserStore struct {
	mu    sync.RWMutex
	users map[string]*UserData // key is user_id
	emails map[string]string    // email -> user_id mapping
}

type UserData struct {
	User         *pb.User
	PasswordHash string
}

// NewUserStore creates a new in-memory user store
func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[string]*UserData),
		emails: make(map[string]string),
	}
}

// CreateUser creates a new user
func (s *UserStore) CreateUser(email, password, name string) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if email already exists
	if _, exists := s.emails[email]; exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	userID := uuid.New().String()
	user := &pb.User{
		UserId:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now().Unix(),
	}

	s.users[userID] = &UserData{
		User:         user,
		PasswordHash: string(hashedPassword),
	}
	s.emails[email] = userID

	return user, nil
}

// AuthenticateUser authenticates a user by email and password
func (s *UserStore) AuthenticateUser(email, password string) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, exists := s.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	userData := s.users[userID]
	if err := bcrypt.CompareHashAndPassword([]byte(userData.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return userData.User, nil
}

// GetUser retrieves a user by ID
func (s *UserStore) GetUser(userID string) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userData, exists := s.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return userData.User, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserStore) GetUserByEmail(email string) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, exists := s.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	return s.users[userID].User, nil
}
