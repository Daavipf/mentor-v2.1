package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
	"mentor-api/internal/infrastructure/security"
)

type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResponse, error)
	Login(ctx context.Context, input LoginInput) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID string) (*UserResponse, error)
}

type authUseCase struct {
	userRepo       repository.UserRepository
	passwordHasher security.PasswordHasher
	tokenManager   security.TokenManager
	tokenTTL       time.Duration
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	passwordHasher security.PasswordHasher,
	tokenManager security.TokenManager,
	tokenTTL time.Duration,
) UseCase {
	return &authUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
		tokenTTL:       tokenTTL,
	}
}

func (u *authUseCase) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	existingUser, err := u.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrUserEmailConflict
	}

	hashedPassword, err := u.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &entity.User{
		ID:        uuid.New().String(),
		Email:     input.Email,
		Password:  hashedPassword,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := u.tokenManager.GenerateToken(user.ID, user.Email, u.tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (u *authUseCase) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := u.passwordHasher.Compare(user.Password, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := u.tokenManager.GenerateToken(user.ID, user.Email, u.tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (u *authUseCase) GetProfile(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

