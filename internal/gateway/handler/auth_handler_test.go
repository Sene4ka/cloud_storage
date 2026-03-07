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

func (m *MockAuthClient) Login(ctx context.Context, in *api.LoginRequest, opts ...grpc.CallOption) (*api.LoginResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.LoginResponse), args.Error(1)
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

func TestAuthHandler_HandleRegister_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Register", mock.Anything, mock.Anything).Return(&api.RegisterResponse{
		UserId:           "user-123",
		Email:            "test@example.com",
		Name:             "Test User",
		AccessToken:      "access-token",
		AccessExpiresIn:  900,
		RefreshToken:     "refresh-token",
		RefreshExpiresIn: 604800,
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

func TestAuthHandler_HandleRegister_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	mockClient.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("user already exists"))

	req := NewTestRequest(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	})
	rr := httptest.NewRecorder()

	handler.HandleRegister(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "user already exists")
	mockClient.AssertExpectations(t)
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

func TestAuthHandler_HandleLogin_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rr := httptest.NewRecorder()

	handler.HandleLogin(rr, req)

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

func TestAuthHandler_HandleRefresh_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/refresh", nil)
	rr := httptest.NewRecorder()

	handler.HandleRefresh(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
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

func TestAuthHandler_HandleLogout_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockClient := new(MockAuthClient)
	handler := NewAuthHandler(mockClient)

	req := NewTestRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
