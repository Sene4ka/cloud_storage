package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
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

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
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

func (m *MockTokenManager) GenerateTempToken(userID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateTempToken(token string) (*utils.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenClaims), args.Error(1)
}

type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) Send2FACode(ctx context.Context, input *api.Send2FACodeRequest, opts ...grpc.CallOption) (*api.Send2FACodeResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Send2FACodeResponse), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@example.com" && u.PasswordHash != ""
	})).Return(nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, 5*time.Minute).Return(nil)

	input := &RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	output, err := svc.Register(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "test@example.com", output.Email)
	assert.Contains(t, output.Message, "Verification code sent")
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockMail.AssertExpectations(t)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

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

func TestAuthService_RegisterComplete_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user := &models.User{
		ID:         "user-123",
		Email:      "test@example.com",
		Name:       "Test User",
		IsVerified: false,
	}

	mockCache.On("Get", mock.Anything, "verify:user-123").Return("123456", nil)
	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.IsVerified == true
	})).Return(nil)
	mockCache.On("Del", mock.Anything, "verify:user-123").Return(nil)
	mockTokenMgr.On("GenerateTokenPair", "user-123", "test@example.com").Return("access-token", "refresh-token", nil)
	mockCache.On("Set", mock.Anything, "refresh:user-123", "refresh-token", 7*24*time.Hour).Return(nil)

	input := &RegisterCompleteInput{
		UserID: "user-123",
		Code:   "123456",
	}

	output, err := svc.RegisterComplete(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "access-token", output.AccessToken)
	assert.Equal(t, "refresh-token", output.RefreshToken)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_RegisterComplete_InvalidCode(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	mockCache.On("Get", mock.Anything, "verify:user-123").Return("123456", nil)

	input := &RegisterCompleteInput{
		UserID: "user-123",
		Code:   "654321",
	}

	output, err := svc.RegisterComplete(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid verification code")
	assert.Nil(t, output)
	mockCache.AssertExpectations(t)
}

func TestAuthService_RegisterComplete_CodeExpired(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	mockCache.On("Get", mock.Anything, "verify:user-123").Return("", errors.New("key not found"))

	input := &RegisterCompleteInput{
		UserID: "user-123",
		Code:   "123456",
	}

	output, err := svc.RegisterComplete(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or expired")
	assert.Nil(t, output)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Login_Success_2FADisabled(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"
	user.IsVerified = true
	user.Is2FAEnabled = false

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
	assert.False(t, output.Requires2FA)
	mockRepo.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Login_Success_2FAEnabled(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"
	user.IsVerified = true
	user.Is2FAEnabled = true

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, "2fa:user-123", mock.Anything, 5*time.Minute).Return(nil)
	mockTokenMgr.On("GenerateTempToken", "user-123", "test@example.com").Return("temp-token", nil)

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := svc.Login(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "user-123", output.UserID)
	assert.Equal(t, "temp-token", output.TempToken)
	assert.True(t, output.Requires2FA)
	assert.Contains(t, output.Message, "2FA code sent")
	mockRepo.AssertExpectations(t)
	mockMail.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockTokenMgr.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

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

func TestAuthService_Login_EmailNotVerified(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.IsVerified = false

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := svc.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email not verified")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_LoginComplete_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	claims := &utils.TokenClaims{
		UserID: "user-123",
		Email:  "test@example.com",
	}

	user := &models.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockTokenMgr.On("ValidateTempToken", "temp-token").Return(claims, nil)
	mockCache.On("Get", mock.Anything, "2fa:user-123").Return("123456", nil)
	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockTokenMgr.On("GenerateTokenPair", "user-123", "test@example.com").Return("access-token", "refresh-token", nil)
	mockCache.On("Set", mock.Anything, "refresh:user-123", "refresh-token", 7*24*time.Hour).Return(nil)
	mockCache.On("Del", mock.Anything, "2fa:user-123").Return(nil)

	input := &LoginCompleteInput{
		TempToken: "temp-token",
		Code:      "123456",
	}

	output, err := svc.LoginComplete(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "access-token", output.AccessToken)
	assert.Equal(t, "refresh-token", output.RefreshToken)
	mockTokenMgr.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_LoginComplete_InvalidCode(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	claims := &utils.TokenClaims{
		UserID: "user-123",
		Email:  "test@example.com",
	}

	mockTokenMgr.On("ValidateTempToken", "temp-token").Return(claims, nil)
	mockCache.On("Get", mock.Anything, "2fa:user-123").Return("123456", nil)

	input := &LoginCompleteInput{
		TempToken: "temp-token",
		Code:      "654321",
	}

	output, err := svc.LoginComplete(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid 2FA code")
	assert.Nil(t, output)
	mockTokenMgr.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Enable2FA_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"
	user.Is2FAEnabled = false

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, "enable_2fa:user-123", mock.Anything, 5*time.Minute).Return(nil)

	input := &Enable2FAInput{
		UserID:   "user-123",
		Password: "password123",
	}

	output, err := svc.Enable2FA(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.Message, "2FA code sent")
	mockRepo.AssertExpectations(t)
	mockMail.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Enable2FA_InvalidPassword(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)

	input := &Enable2FAInput{
		UserID:   "user-123",
		Password: "wrongpassword",
	}

	output, err := svc.Enable2FA(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Enable2FAComplete_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user := &models.User{
		ID:           "user-123",
		Email:        "test@example.com",
		Is2FAEnabled: false,
	}

	mockCache.On("Get", mock.Anything, "enable_2fa:user-123").Return("123456", nil)
	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Is2FAEnabled == true
	})).Return(nil)
	mockCache.On("Del", mock.Anything, "enable_2fa:user-123").Return(nil)

	input := &Enable2FACompleteInput{
		UserID: "user-123",
		Code:   "123456",
	}

	output, err := svc.Enable2FAComplete(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Is2FAEnabled)
	assert.Contains(t, output.Message, "enabled successfully")
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Disable2FA_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"
	user.Is2FAEnabled = true

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, "disable_2fa:user-123", mock.Anything, 5*time.Minute).Return(nil)

	input := &Disable2FAInput{
		UserID:   "user-123",
		Password: "password123",
	}

	output, err := svc.Disable2FA(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.Message, "Verification code sent")
	mockRepo.AssertExpectations(t)
	mockMail.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_Disable2FA_NotEnabled(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"
	user.Is2FAEnabled = false

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)

	input := &Disable2FAInput{
		UserID:   "user-123",
		Password: "password123",
	}

	output, err := svc.Disable2FA(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "2FA is not enabled")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Disable2FAComplete_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user := &models.User{
		ID:           "user-123",
		Email:        "test@example.com",
		Is2FAEnabled: true,
	}

	mockCache.On("Get", mock.Anything, "disable_2fa:user-123").Return("123456", nil)
	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Is2FAEnabled == false
	})).Return(nil)
	mockCache.On("Del", mock.Anything, "disable_2fa:user-123").Return(nil)

	input := &Disable2FACompleteInput{
		UserID: "user-123",
		Code:   "123456",
	}

	output, err := svc.Disable2FAComplete(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.Is2FAEnabled)
	assert.Contains(t, output.Message, "disabled successfully")
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangeEmail_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(false, nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, "change_email:user-123", mock.Anything, 5*time.Minute).Return(nil)

	input := &ChangeEmailInput{
		UserID:          "user-123",
		CurrentPassword: "password123",
		NewEmail:        "new@example.com",
	}

	output, err := svc.ChangeEmail(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.Message, "Verification code sent")
	mockRepo.AssertExpectations(t)
	mockMail.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_ChangeEmail_EmailInUse(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(true, nil)

	input := &ChangeEmailInput{
		UserID:          "user-123",
		CurrentPassword: "password123",
		NewEmail:        "new@example.com",
	}

	output, err := svc.ChangeEmail(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already in use")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user, _ := models.NewUser("test@example.com", "password123", "Test User")
	user.ID = "user-123"

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockMail.On("Send2FACode", mock.Anything, mock.Anything).Return(&api.Send2FACodeResponse{
		Success: true,
	}, nil)
	mockCache.On("Set", mock.Anything, "change_password:user-123", mock.Anything, 5*time.Minute).Return(nil)

	input := &ChangePasswordInput{
		UserID:          "user-123",
		CurrentPassword: "password123",
		NewPassword:     "newpassword123",
	}

	output, err := svc.ChangePassword(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.Message, "Verification code sent")
	mockRepo.AssertExpectations(t)
	mockMail.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAuthService_ChangeMeta_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

	user := &models.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Old Name",
	}

	mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Name == "New Name"
	})).Return(nil)

	input := &ChangeMetaInput{
		UserID: "user-123",
		Name:   "New Name",
	}

	output, err := svc.ChangeMeta(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "New Name", output.Name)
	assert.Contains(t, output.Message, "Profile updated")
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

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

func TestAuthService_ValidateToken_Valid(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

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

func TestAuthService_Logout_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepository)
	mockCache := new(MockTokenCache)
	mockTokenMgr := new(MockTokenManager)
	mockMail := new(MockMailService)
	config := &configs.Config{
		JWT: configs.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(mockRepo, mockCache, mockTokenMgr, mockMail, config)

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
