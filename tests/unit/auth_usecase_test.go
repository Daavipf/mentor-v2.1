package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mentor-api/internal/domain"
	"mentor-api/internal/infrastructure/security"
	"mentor-api/internal/usecase/auth"
	"mentor-api/tests/mocks"
)

func setupAuthUseCase() auth.UseCase {
	userRepo := mocks.NewMockUserRepository()
	hasher := security.NewBcryptHasher(4)
	tokenMgr := security.NewJWTManager("test-secret-key-1234567890123456")
	return auth.NewAuthUseCase(userRepo, hasher, tokenMgr, 1*time.Hour)
}

func TestAuth_Register_Success(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	resp, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.NotEmpty(t, resp.User.ID)
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	_, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "First",
		Email:    "dupe@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = uc.Register(ctx, auth.RegisterInput{
		Name:     "Second",
		Email:    "dupe@example.com",
		Password: "password456",
	})
	assert.ErrorIs(t, err, domain.ErrUserEmailConflict)
}

func TestAuth_Register_EmptyInput(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	_, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "",
		Email:    "empty@example.com",
		Password: "pass",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestAuth_Login_Success(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	_, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "Login User",
		Email:    "login@example.com",
		Password: "securepassword",
	})
	require.NoError(t, err)

	loginResp, err := uc.Login(ctx, auth.LoginInput{
		Email:    "login@example.com",
		Password: "securepassword",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)
	assert.Equal(t, "login@example.com", loginResp.User.Email)
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	_, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "Wrong Pass",
		Email:    "wrongpass@example.com",
		Password: "correctpassword",
	})
	require.NoError(t, err)

	_, err = uc.Login(ctx, auth.LoginInput{
		Email:    "wrongpass@example.com",
		Password: "wrongpassword",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuth_GetProfile(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	regResp, err := uc.Register(ctx, auth.RegisterInput{
		Name:     "Profile User",
		Email:    "profile@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	profile, err := uc.GetProfile(ctx, regResp.User.ID)
	require.NoError(t, err)
	assert.Equal(t, regResp.User.ID, profile.ID)
	assert.Equal(t, "profile@example.com", profile.Email)
}

