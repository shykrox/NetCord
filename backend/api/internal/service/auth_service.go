package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"netcord/backend/api/internal/auth"
	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	Token string            `json:"token"`
	User  models.PublicUser `json:"user"`
}

type AuthService struct {
	users  repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = normalizeEmail(input.Email)

	if fields := validateRegisterInput(input); len(fields) > 0 {
		return AuthResult{}, &ValidationError{Fields: fields}
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	displayName := input.Username
	user := models.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
		DisplayName:  &displayName,
		Status:       "offline",
	}

	created, err := s.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserConflict) {
			return AuthResult{}, &ValidationError{
				Fields: map[string]string{
					"account": "username or email already in use",
				},
			}
		}
		return AuthResult{}, err
	}

	token, err := s.tokens.Generate(created.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{Token: token, User: created.Public()}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	input.Email = normalizeEmail(input.Email)
	if input.Email == "" || input.Password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if !auth.CheckPassword(user.PasswordHash, input.Password) {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{Token: token, User: user.Public()}, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (models.PublicUser, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return models.PublicUser{}, err
	}
	return user.Public(), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func utcNow() time.Time {
	return time.Now().UTC()
}
