package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Register(ctx context.Context, in *api.RegisterRequest, opts ...grpc.CallOption) (*api.RegisterResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.RegisterResponse), args.Error(1)
}

func (m *MockAuthClient) RegisterComplete(ctx context.Context, in *api.RegisterCompleteRequest, opts ...grpc.CallOption) (*api.RegisterCompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.RegisterCompleteResponse), args.Error(1)
}

func (m *MockAuthClient) Login(ctx context.Context, in *api.LoginRequest, opts ...grpc.CallOption) (*api.LoginResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.LoginResponse), args.Error(1)
}

func (m *MockAuthClient) LoginComplete(ctx context.Context, in *api.LoginCompleteRequest, opts ...grpc.CallOption) (*api.LoginCompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.LoginCompleteResponse), args.Error(1)
}

func (m *MockAuthClient) Refresh(ctx context.Context, in *api.RefreshRequest, opts ...grpc.CallOption) (*api.RefreshResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.RefreshResponse), args.Error(1)
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, in *api.ValidateTokenRequest, opts ...grpc.CallOption) (*api.ValidateTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ValidateTokenResponse), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context, in *api.LogoutRequest, opts ...grpc.CallOption) (*api.LogoutResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.LogoutResponse), args.Error(1)
}

func (m *MockAuthClient) Enable2FA(ctx context.Context, in *api.Enable2FARequest, opts ...grpc.CallOption) (*api.Enable2FAResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Enable2FAResponse), args.Error(1)
}

func (m *MockAuthClient) Enable2FAComplete(ctx context.Context, in *api.Enable2FACompleteRequest, opts ...grpc.CallOption) (*api.Enable2FACompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Enable2FACompleteResponse), args.Error(1)
}

func (m *MockAuthClient) Disable2FA(ctx context.Context, in *api.Disable2FARequest, opts ...grpc.CallOption) (*api.Disable2FAResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Disable2FAResponse), args.Error(1)
}

func (m *MockAuthClient) Disable2FAComplete(ctx context.Context, in *api.Disable2FACompleteRequest, opts ...grpc.CallOption) (*api.Disable2FACompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Disable2FACompleteResponse), args.Error(1)
}

func (m *MockAuthClient) ChangeEmail(ctx context.Context, in *api.ChangeEmailRequest, opts ...grpc.CallOption) (*api.ChangeEmailResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ChangeEmailResponse), args.Error(1)
}

func (m *MockAuthClient) ChangeEmailComplete(ctx context.Context, in *api.ChangeEmailCompleteRequest, opts ...grpc.CallOption) (*api.ChangeEmailCompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ChangeEmailCompleteResponse), args.Error(1)
}

func (m *MockAuthClient) ChangePassword(ctx context.Context, in *api.ChangePasswordRequest, opts ...grpc.CallOption) (*api.ChangePasswordResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ChangePasswordResponse), args.Error(1)
}

func (m *MockAuthClient) ChangePasswordComplete(ctx context.Context, in *api.ChangePasswordCompleteRequest, opts ...grpc.CallOption) (*api.ChangePasswordCompleteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ChangePasswordCompleteResponse), args.Error(1)
}

func (m *MockAuthClient) ChangeMeta(ctx context.Context, in *api.ChangeMetaRequest, opts ...grpc.CallOption) (*api.ChangeMetaResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ChangeMetaResponse), args.Error(1)
}

func TestAuthHandler_HandleRegister_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Register", mock.Anything, mock.Anything).Return(&api.RegisterResponse{
		UserId:  "user-123",
		Email:   "test@example.com",
		Name:    "Test User",
		Message: "Verification code sent to email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	})
	rr := httptest.NewRecorder()

	handler.HandleRegister(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "user-123")
	assert.Contains(t, rr.Body.String(), "test@example.com")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleRegister_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
	rr := httptest.NewRecorder()

	handler.HandleRegister(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_HandleRegister_InvalidBody(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleRegister(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid request body")
}

func TestAuthHandler_HandleRegisterComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("RegisterComplete", mock.Anything, mock.Anything).Return(&api.RegisterCompleteResponse{
		UserId:           "user-123",
		Email:            "test@example.com",
		Name:             "Test User",
		AccessToken:      "access-token",
		AccessExpiresIn:  900,
		RefreshToken:     "refresh-token",
		RefreshExpiresIn: 604800,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/register/complete", map[string]string{
		"user_id": "user-123",
		"code":    "123456",
	})
	rr := httptest.NewRecorder()

	handler.HandleRegisterComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "access-token")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleRegisterComplete_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/register/complete", nil)
	rr := httptest.NewRecorder()

	handler.HandleRegisterComplete(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_HandleLogin_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Login", mock.Anything, mock.Anything).Return(&api.LoginResponse{
		UserId:           "user-123",
		Email:            "test@example.com",
		Name:             "Test User",
		AccessToken:      "access-token",
		AccessExpiresIn:  900,
		RefreshToken:     "refresh-token",
		RefreshExpiresIn: 604800,
		Requires_2Fa:     false,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})
	rr := httptest.NewRecorder()

	handler.HandleLogin(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "access-token")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLogin_Requires2FA(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Login", mock.Anything, mock.Anything).Return(&api.LoginResponse{
		UserId:       "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		TempToken:    "temp-token",
		Requires_2Fa: true,
		Message:      "2FA code sent to email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})
	rr := httptest.NewRecorder()

	handler.HandleLogin(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "temp-token")
	assert.Contains(t, rr.Body.String(), "true")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLogin_InvalidCredentials(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Login", mock.Anything, mock.Anything).Return(nil, errors.New("invalid credentials"))

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})
	rr := httptest.NewRecorder()

	handler.HandleLogin(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLoginComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("LoginComplete", mock.Anything, mock.Anything).Return(&api.LoginCompleteResponse{
		UserId:           "user-123",
		Email:            "test@example.com",
		Name:             "Test User",
		AccessToken:      "access-token",
		AccessExpiresIn:  900,
		RefreshToken:     "refresh-token",
		RefreshExpiresIn: 604800,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/login/complete", map[string]string{
		"temp_token": "temp-token",
		"code":       "123456",
	})
	rr := httptest.NewRecorder()

	handler.HandleLoginComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "access-token")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLoginComplete_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/login/complete", nil)
	rr := httptest.NewRecorder()

	handler.HandleLoginComplete(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_HandleRefresh_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Refresh", mock.Anything, mock.Anything).Return(&api.RefreshResponse{
		AccessToken:      "new-access-token",
		AccessExpiresIn:  900,
		RefreshToken:     "new-refresh-token",
		RefreshExpiresIn: 604800,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "refresh-token",
	})
	rr := httptest.NewRecorder()

	handler.HandleRefresh(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "new-access-token")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLogout_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Logout", mock.Anything, mock.Anything).Return(&api.LogoutResponse{
		Success: true,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": "refresh-token",
	})
	rr := httptest.NewRecorder()

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "true")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleLogout_Failed(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Logout", mock.Anything, mock.Anything).Return(&api.LogoutResponse{
		Success: false,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": "invalid-token",
	})
	rr := httptest.NewRecorder()

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "logout failed")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleEnable2FA_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Enable2FA", mock.Anything, mock.Anything).Return(&api.Enable2FAResponse{
		Message: "2FA code sent to email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/enable", map[string]string{
		"password": "password123",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleEnable2FA(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "2FA code sent")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleEnable2FA_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/2fa/enable", nil)
	rr := httptest.NewRecorder()

	handler.HandleEnable2FA(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_HandleEnable2FA_NoAuth(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/enable", map[string]string{
		"password": "password123",
	})
	rr := httptest.NewRecorder()

	handler.HandleEnable2FA(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "user_id not found in context")
}

func TestAuthHandler_HandleEnable2FAComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Enable2FAComplete", mock.Anything, mock.Anything).Return(&api.Enable2FACompleteResponse{
		Is_2FaEnabled: true,
		Message:       "2FA enabled successfully",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/enable/complete", map[string]string{
		"code": "123456",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleEnable2FAComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "2FA enabled successfully")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleDisable2FA_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Disable2FA", mock.Anything, mock.Anything).Return(&api.Disable2FAResponse{
		Message: "Verification code sent to email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/disable", map[string]string{
		"password": "password123",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleDisable2FA(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Verification code sent")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleDisable2FA_NoAuth(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/disable", map[string]string{
		"password": "password123",
	})
	rr := httptest.NewRecorder()

	handler.HandleDisable2FA(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "user_id not found in context")
}

func TestAuthHandler_HandleDisable2FAComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Disable2FAComplete", mock.Anything, mock.Anything).Return(&api.Disable2FACompleteResponse{
		Is_2FaEnabled: false,
		Message:       "2FA disabled successfully",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/2fa/disable/complete", map[string]string{
		"code": "123456",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleDisable2FAComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "2FA disabled successfully")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangeEmail_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("ChangeEmail", mock.Anything, mock.Anything).Return(&api.ChangeEmailResponse{
		Message: "Verification code sent to new email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/email/change", map[string]string{
		"current_password": "password123",
		"new_email":        "new@example.com",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleChangeEmail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Verification code sent")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangeEmail_NoAuth(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/email/change", map[string]string{
		"current_password": "password123",
		"new_email":        "new@example.com",
	})
	rr := httptest.NewRecorder()

	handler.HandleChangeEmail(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "user_id not found in context")
}

func TestAuthHandler_HandleChangeEmailComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("ChangeEmailComplete", mock.Anything, mock.Anything).Return(&api.ChangeEmailCompleteResponse{
		Email:   "new@example.com",
		Message: "Email changed successfully",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/email/change/complete", map[string]string{
		"code": "123456",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleChangeEmailComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "new@example.com")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangePassword_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("ChangePassword", mock.Anything, mock.Anything).Return(&api.ChangePasswordResponse{
		Message: "Verification code sent to email",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/password/change", map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword123",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleChangePassword(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Verification code sent")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangePassword_NoAuth(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/password/change", map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword123",
	})
	rr := httptest.NewRecorder()

	handler.HandleChangePassword(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "user_id not found in context")
}

func TestAuthHandler_HandleChangePasswordComplete_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("ChangePasswordComplete", mock.Anything, mock.Anything).Return(&api.ChangePasswordCompleteResponse{
		Message: "Password changed successfully",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/password/change/complete", map[string]string{
		"code": "123456",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleChangePasswordComplete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Password changed")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangeMeta_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("ChangeMeta", mock.Anything, mock.Anything).Return(&api.ChangeMetaResponse{
		UserId:  "user-123",
		Name:    "New Name",
		Message: "Profile updated successfully",
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/meta/change", map[string]string{
		"name": "New Name",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleChangeMeta(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "New Name")
	mockClient.AssertExpectations(t)
}

func TestAuthHandler_HandleChangeMeta_NoAuth(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/meta/change", map[string]string{
		"name": "New Name",
	})
	rr := httptest.NewRecorder()

	handler.HandleChangeMeta(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "user_id not found in context")
}
