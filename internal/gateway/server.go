package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	config         *configs.Config
	authClient     api.AuthServiceClient
	metadataClient api.MetadataServiceClient
	fileClient     api.FileServiceClient
	httpServer     *http.Server
}

func NewServer(config *configs.Config) (*Server, error) {
	authCC, err := grpc.NewClient(config.Services.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("auth conn: %w", err)
	}
	authConn := grpc.ClientConnInterface(authCC)

	metadataCC, err := grpc.NewClient(config.Services.MetadataAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("metadata conn: %w", err)
	}
	metadataConn := grpc.ClientConnInterface(metadataCC)

	fileCC, err := grpc.NewClient(config.Services.FileAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("file conn: %w", err)
	}
	fileConn := grpc.ClientConnInterface(fileCC)

	server := &Server{
		config:         config,
		authClient:     api.NewAuthServiceClient(authConn),
		metadataClient: api.NewMetadataServiceClient(metadataConn),
		fileClient:     api.NewFileServiceClient(fileConn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/api/v1/auth/register", server.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", server.handleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", server.handleRefresh)
	mux.HandleFunc("/api/v1/auth/logout", server.withAuth(server.handleLogout))
	mux.HandleFunc("/api/v1/files", server.withAuth(server.handleFiles))
	mux.HandleFunc("/api/v1/files/", server.withAuth(server.handleFileDetail))
	mux.HandleFunc("/api/v1/files/upload", server.withAuth(server.handleInitiateUpload))
	mux.HandleFunc("/api/v1/files/upload/complete", server.withAuth(server.handleCompleteUpload))
	mux.HandleFunc("/api/v1/files/download/", server.withAuth(server.handleDownloadLink))
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port),
		Handler:      server.corsMiddleware(mux),
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	return server, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		resp, err := s.authClient.ValidateToken(r.Context(), &api.ValidateTokenRequest{Token: token})
		if err != nil || !resp.Valid {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", resp.UserId)
		ctx = context.WithValue(ctx, "email", resp.Email)
		ctx = context.WithValue(ctx, "token", token)

		next(w, r.WithContext(ctx))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := s.authClient.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, resp)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := s.authClient.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusUnauthorized)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := s.authClient.Refresh(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "refresh failed"}`, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req api.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := s.authClient.Logout(r.Context(), &req)
	if err != nil || !resp.Success {
		http.Error(w, `{"error": "logout failed"}`, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	switch r.Method {
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		resp, err := s.metadataClient.ListMetadata(r.Context(), &api.ListMetadataRequest{
			UserId:    userID,
			Page:      int32(page),
			PageSize:  int32(pageSize),
			SortBy:    r.URL.Query().Get("sort_by"),
			SortOrder: r.URL.Query().Get("sort_order"),
			Search:    r.URL.Query().Get("search"),
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, http.StatusOK, resp)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFileDetail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	switch r.Method {
	case http.MethodGet:
		resp, err := s.metadataClient.GetMetadata(r.Context(), &api.GetMetadataRequest{
			Id:     fileID,
			UserId: userID,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		jsonResponse(w, http.StatusOK, resp)
	case http.MethodPut:
		var req api.UpdateMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		req.Id = fileID
		req.UserId = userID
		resp, err := s.metadataClient.UpdateMetadata(r.Context(), &req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, http.StatusOK, resp)
	case http.MethodDelete:
		_, err := s.metadataClient.DeleteMetadata(r.Context(), &api.DeleteMetadataRequest{
			Id:     fileID,
			UserId: userID,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, http.StatusOK, map[string]bool{"success": true})
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleInitiateUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	var req api.InitiateUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.UserId = userID
	resp, err := s.fileClient.InitiateUpload(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", resp.UploadUrl)
	jsonResponse(w, http.StatusCreated, resp)
}

func (s *Server) handleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	var req api.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.UserId = userID
	resp, err := s.fileClient.CompleteUpload(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handleDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/download/")
	expiresIn := int64(3600) // 1 hour default
	if exp := r.URL.Query().Get("expires_in"); exp != "" {
		if val, err := strconv.ParseInt(exp, 10, 64); err == nil {
			expiresIn = val
		}
	}

	resp, err := s.fileClient.GetDownloadLink(r.Context(), &api.GetDownloadLinkRequest{
		FileId:    fileID,
		UserId:    userID,
		ExpiresIn: expiresIn,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, resp)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
