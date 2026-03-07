package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

type MockTokenCache struct {
	mock.Mock
}

func (m *MockTokenCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockTokenCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockTokenCache) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockTokenCache) Exists(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateTokenPair(userID, email string) (string, string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenManager) ValidateAccessToken(token string) (*utils.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenClaims), args.Error(1)
}

func (m *MockTokenManager) ValidateRefreshToken(token string) (*utils.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenClaims), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@example.com" && u.PasswordHash != ""
	})).Return(nil)
	mockTokenMgr.On("GenerateTokenPair", mock.Anything, "test@example.com").Return("access-token", "refresh-token", nil)
	mockCache.On("Set", mock.Anything, mock.Anything, "refresh-token", 7*24*time.Hour).Return(nil)

	input := &RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	output, err := svc.Register(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "test@example.com", output.Email)
	assert.Equal(t, "access-token", output.AccessToken)
	assert.Equal(t, "refresh-token", output.RefreshToken)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(true, nil)

	input := &RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	output, err := svc.Register(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockTokenMgr.On("GenerateTokenPair", "user-123", "test@example.com").Return("access-token", "refresh-token", nil)
	mockCache.On("Set", mock.Anything, "refresh:user-123", "refresh-token", 7*24*time.Hour).Return(nil)

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := svc.Login(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "user-123", output.UserID)
	assert.Equal(t, "access-token", output.AccessToken)
	mockRepo.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	output, err := svc.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("not found"))

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := svc.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	claims := &utils.TokenClaims{
		UserID: "user-123",
		Email:  "test@example.com",
	}

	mockCache.On("Exists", mock.Anything, "blacklist:refresh-token").Return(int64(0), nil)
	mockTokenMgr.On("ValidateRefreshToken", "refresh-token").Return(claims, nil)
	mockCache.On("Get", mock.Anything, "refresh:user-123").Return("refresh-token", nil)
	mockTokenMgr.On("GenerateTokenPair", "user-123", "test@example.com").Return("new-access", "new-refresh", nil)
	mockCache.On("Set", mock.Anything, "refresh:user-123", "new-refresh", 7*24*time.Hour).Return(nil)

	input := &RefreshInput{
		RefreshToken: "refresh-token",
	}

	output, err := svc.Refresh(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "new-access", output.AccessToken)
	assert.Equal(t, "new-refresh", output.RefreshToken)
	mockCache.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_Refresh_TokenBlacklisted(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockCache.On("Exists", mock.Anything, "blacklist:refresh-token").Return(int64(1), nil)

	input := &RefreshInput{
		RefreshToken: "refresh-token",
	}

	output, err := svc.Refresh(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blacklisted")
	assert.Nil(t, output)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockCache.On("Exists", mock.Anything, "blacklist:refresh-token").Return(int64(0), nil)
	mockTokenMgr.On("ValidateRefreshToken", "refresh-token").Return(nil, errors.New("invalid"))

	input := &RefreshInput{
		RefreshToken: "refresh-token",
	}

	output, err := svc.Refresh(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
	assert.Nil(t, output)
	mockCache.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Valid(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	claims := &utils.TokenClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	mockTokenMgr.On("ValidateAccessToken", "valid-token").Return(claims, nil)

	input := &ValidateTokenInput{
		Token: "valid-token",
	}

	output, err := svc.ValidateToken(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Valid)
	assert.Equal(t, "user-123", output.UserID)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockTokenMgr.On("ValidateAccessToken", "invalid-token").Return(nil, errors.New("invalid"))

	input := &ValidateTokenInput{
		Token: "invalid-token",
	}

	output, err := svc.ValidateToken(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.Valid)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_Logout_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	claims := &utils.TokenClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	mockTokenMgr.On("ValidateRefreshToken", "refresh-token").Return(claims, nil)
	mockCache.On("Set", mock.Anything, "blacklist:refresh-token", "1", mock.Anything).Return(nil)
	mockCache.On("Del", mock.Anything, "refresh:user-123").Return(nil)

	input := &LogoutInput{
		RefreshToken: "refresh-token",
	}

	output, err := svc.Logout(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)
	mockTokenMgr.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, config)

	mockTokenMgr.On("ValidateRefreshToken", "invalid-token").Return(nil, errors.New("invalid"))

	input := &LogoutInput{
		RefreshToken: "invalid-token",
	}

	output, err := svc.Logout(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.Success)
	mockTokenMgr.AssertExpectations(t)
}
