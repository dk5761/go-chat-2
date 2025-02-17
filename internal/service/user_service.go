package service

import (
	"context"
	"errors"
	"time"

	"github.com/chat-backend/internal/models"
	"github.com/chat-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc"
	"golang.org/x/crypto/bcrypt"
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
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
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
		return nil, errors.New("user with this email already exists")
	}

	if _, err := s.userRepo.GetByUsername(ctx, input.Username); err == nil {
		return nil, errors.New("username is already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Update user status
	var wg conc.WaitGroup
	wg.Go(func() {
		if err := s.statusRepo.UpdateStatus(ctx, user.ID, "online"); err != nil {
			// Log error but don't fail the registration
			// TODO: Add proper logging
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
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Update last seen and status concurrently
	var wg conc.WaitGroup
	wg.Go(func() {
		if err := s.userRepo.UpdateLastSeen(ctx, user.ID); err != nil {
			// Log error but don't fail the login
			// TODO: Add proper logging
		}
	})

	wg.Go(func() {
		if err := s.statusRepo.UpdateStatus(ctx, user.ID, "online"); err != nil {
			// Log error but don't fail the login
			// TODO: Add proper logging
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
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
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
		return err
	}
	return s.statusRepo.UpdateStatus(ctx, userUUID, status)
}

func (s *UserService) GetUserStatus(ctx context.Context, userID string) (string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	return s.statusRepo.GetStatus(ctx, userUUID)
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
