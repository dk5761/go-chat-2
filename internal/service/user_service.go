package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/chat-backend/internal/apperrors"
	"github.com/chat-backend/internal/models"
	"github.com/chat-backend/internal/repository"
)

type UserService struct {
	userRepo   repository.UserRepository
	statusRepo repository.StatusRepository
	jwtSecret  []byte
}

func NewUserService(userRepo repository.UserRepository, statusRepo repository.StatusRepository, jwtSecret string) *UserService {
	return &UserService{
		userRepo:   userRepo,
		statusRepo: statusRepo,
		jwtSecret:  []byte(jwtSecret),
	}
}

type RegisterUserInput struct {
	Username string `json:"username" binding:"required" example:"johndoe" msg:"Username is required"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com" msg:"Please enter a valid email address"`
	Password string `json:"password" binding:"required,min=8" example:"securepass123" msg:"Password must be at least 8 characters long"`
	FullName string `json:"full_name" binding:"required" example:"John Doe" msg:"Full name is required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*AuthResponse, error) {
	// Check if user exists
	if _, err := s.userRepo.GetByEmail(ctx, input.Email); err == nil {
		return nil, apperrors.ErrEmailExists
	}

	if _, err := s.userRepo.GetByUsername(ctx, input.Username); err == nil {
		return nil, apperrors.ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrServerError
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashedPassword),
		FullName:  input.FullName,
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.ErrServerError
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, apperrors.ErrServerError
	}

	// Update user status
	var wg conc.WaitGroup
	wg.Go(func() {
		bgCtx := context.Background()
		if err := s.statusRepo.UpdateStatus(bgCtx, user.ID, "online"); err != nil {
			logrus.Error("Failed to update user status: ", err)
		}
	})

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, apperrors.ErrServerError
	}

	// Update last seen and status concurrently
	var wg conc.WaitGroup
	wg.Go(func() {
		bgCtx := context.Background()
		if err := s.userRepo.UpdateLastSeen(bgCtx, user.ID); err != nil {
			logrus.Error("Failed to update user last seen: ", err)
		}
	})

	wg.Go(func() {
		bgCtx := context.Background()
		if err := s.statusRepo.UpdateStatus(bgCtx, user.ID, "online"); err != nil {
			logrus.Error("Failed to update user status: ", err)
		}
	})

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return apperrors.ErrInvalidPassword
	}

	if len(newPassword) < 8 {
		return apperrors.ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.ErrServerError
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperrors.ErrServerError
	}

	return nil
}

func (s *UserService) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *UserService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims["user_id"].(string))
		if err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token")
}

func (s *UserService) UpdateStatus(ctx context.Context, userID string, status string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Use a background context with timeout for status updates
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.statusRepo.UpdateStatus(timeoutCtx, userUUID, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (s *UserService) GetUserStatus(ctx context.Context, userID string) (string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Use a background context with timeout for status retrieval
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := s.statusRepo.GetStatus(timeoutCtx, userUUID)
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}
	return status, nil
}

func (s *UserService) GetMultiUserStatus(ctx context.Context, userIDs []string) (map[string]string, error) {
	uuids := make([]uuid.UUID, len(userIDs))
	for i, id := range userIDs {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids[i] = uuid
	}

	statuses, err := s.statusRepo.GetMultiStatus(ctx, uuids)
	if err != nil {
		return nil, err
	}

	// Convert UUID keys to strings
	result := make(map[string]string, len(statuses))
	for id, status := range statuses {
		result[id.String()] = status
	}

	return result, nil
}
