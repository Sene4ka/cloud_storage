package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"google.golang.org/grpc"
)

type AuthClient interface {
	Register(ctx context.Context, in *api.RegisterRequest, opts ...grpc.CallOption) (*api.RegisterResponse, error)
	RegisterComplete(ctx context.Context, in *api.RegisterCompleteRequest, opts ...grpc.CallOption) (*api.RegisterCompleteResponse, error)
	Login(ctx context.Context, in *api.LoginRequest, opts ...grpc.CallOption) (*api.LoginResponse, error)
	LoginComplete(ctx context.Context, in *api.LoginCompleteRequest, opts ...grpc.CallOption) (*api.LoginCompleteResponse, error)
	Refresh(ctx context.Context, in *api.RefreshRequest, opts ...grpc.CallOption) (*api.RefreshResponse, error)
	ValidateToken(ctx context.Context, in *api.ValidateTokenRequest, opts ...grpc.CallOption) (*api.ValidateTokenResponse, error)
	Logout(ctx context.Context, in *api.LogoutRequest, opts ...grpc.CallOption) (*api.LogoutResponse, error)
	Enable2FA(ctx context.Context, in *api.Enable2FARequest, opts ...grpc.CallOption) (*api.Enable2FAResponse, error)
	Enable2FAComplete(ctx context.Context, in *api.Enable2FACompleteRequest, opts ...grpc.CallOption) (*api.Enable2FACompleteResponse, error)
	Disable2FA(ctx context.Context, in *api.Disable2FARequest, opts ...grpc.CallOption) (*api.Disable2FAResponse, error)
	Disable2FAComplete(ctx context.Context, in *api.Disable2FACompleteRequest, opts ...grpc.CallOption) (*api.Disable2FACompleteResponse, error)
	ChangeEmail(ctx context.Context, in *api.ChangeEmailRequest, opts ...grpc.CallOption) (*api.ChangeEmailResponse, error)
	ChangeEmailComplete(ctx context.Context, in *api.ChangeEmailCompleteRequest, opts ...grpc.CallOption) (*api.ChangeEmailCompleteResponse, error)
	ChangePassword(ctx context.Context, in *api.ChangePasswordRequest, opts ...grpc.CallOption) (*api.ChangePasswordResponse, error)
	ChangePasswordComplete(ctx context.Context, in *api.ChangePasswordCompleteRequest, opts ...grpc.CallOption) (*api.ChangePasswordCompleteResponse, error)
	ChangeMeta(ctx context.Context, in *api.ChangeMetaRequest, opts ...grpc.CallOption) (*api.ChangeMetaResponse, error)
}

type AuthHandler struct {
	authClient AuthClient
}

func NewAuthHandler(client AuthClient) *AuthHandler {
	return &AuthHandler{authClient: client}
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusCreated, resp)
}

func (h *AuthHandler) HandleRegisterComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.RegisterCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.RegisterComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusUnauthorized)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleLoginComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.LoginCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.LoginComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Refresh(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "refresh failed"}`, http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Logout(r.Context(), &req)
	if err != nil || !resp.Success {
		http.Error(w, `{"error": "logout failed"}`, http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleEnable2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.Enable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Enable2FA(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleEnable2FAComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.Enable2FACompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Enable2FAComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleDisable2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.Disable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Disable2FA(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleDisable2FAComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.Disable2FACompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Disable2FAComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleChangeEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.ChangeEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.ChangeEmail(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleChangeEmailComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.ChangeEmailCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.ChangeEmailComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.ChangePassword(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleChangePasswordComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.ChangePasswordCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.ChangePasswordComplete(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleChangeMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.ChangeMetaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.ChangeMeta(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}
