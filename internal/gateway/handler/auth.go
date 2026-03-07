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
	Login(ctx context.Context, in *api.LoginRequest, opts ...grpc.CallOption) (*api.LoginResponse, error)
	Refresh(ctx context.Context, in *api.RefreshRequest, opts ...grpc.CallOption) (*api.RefreshResponse, error)
	Logout(ctx context.Context, in *api.LogoutRequest, opts ...grpc.CallOption) (*api.LogoutResponse, error)
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
