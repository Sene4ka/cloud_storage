package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(ctx context.Context, in *api.ValidateTokenRequest, opts ...grpc.CallOption) (*api.ValidateTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ValidateTokenResponse), args.Error(1)
}

func TestCORS(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := CORS(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	corsHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORS_Preflight(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := CORS(handler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rr := httptest.NewRecorder()

	corsHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWithAuth_Success(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockTokenValidator)

	mockValidator.On("ValidateToken", mock.Anything, mock.Anything).Return(&api.ValidateTokenResponse{
		Valid:     true,
		UserId:    "user-123",
		Email:     "test@example.com",
		ExpiresIn: 900,
	}, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		assert.Equal(t, "user-123", userID)
		w.WriteHeader(http.StatusOK)
	})

	authHandler := WithAuth(handler, mockValidator)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockValidator.AssertExpectations(t)
}

func TestWithAuth_NoToken(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockTokenValidator)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := WithAuth(handler, mockValidator)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "authorization header required")
}

func TestWithAuth_InvalidToken(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockTokenValidator)

	mockValidator.On("ValidateToken", mock.Anything, mock.Anything).Return(&api.ValidateTokenResponse{
		Valid: false,
	}, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := WithAuth(handler, mockValidator)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid or expired token")
	mockValidator.AssertExpectations(t)
}

func TestWithAuth_TokenError(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockTokenValidator)

	mockValidator.On("ValidateToken", mock.Anything, mock.Anything).Return(nil, errors.New("validation error"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := WithAuth(handler, mockValidator)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockValidator.AssertExpectations(t)
}
