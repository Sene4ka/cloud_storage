package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/gateway/handler"
	"github.com/Sene4ka/cloud_storage/internal/gateway/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	config      *configs.Config
	httpServer  *http.Server
	authHandler *handler.AuthHandler
	fileHandler *handler.FileHandler
}

func NewServer(config *configs.Config) (*Server, error) {
	authCC, err := grpc.NewClient(config.Services.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("auth conn: %w", err)
	}
	authConn := grpc.ClientConnInterface(authCC)
	authClient := api.NewAuthServiceClient(authConn)

	metadataCC, err := grpc.NewClient(config.Services.MetadataAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("metadata conn: %w", err)
	}
	metadataConn := grpc.ClientConnInterface(metadataCC)
	metadataCLient := api.NewMetadataServiceClient(metadataConn)

	fileCC, err := grpc.NewClient(config.Services.FileAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("file conn: %w", err)
	}
	fileConn := grpc.ClientConnInterface(fileCC)
	fileClient := api.NewFileServiceClient(fileConn)

	server := &Server{
		config:      config,
		authHandler: handler.NewAuthHandler(authClient),
		fileHandler: handler.NewFileHandler(metadataCLient, fileClient),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", server.handleHealth)

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/docs/", server.handleDocs)
	mux.HandleFunc("/docs/swagger/", server.handleSwagger)

	mux.HandleFunc("/api/v1/auth/register", server.authHandler.HandleRegister)
	mux.HandleFunc("/api/v1/auth/register/complete", server.authHandler.HandleRegisterComplete)
	mux.HandleFunc("/api/v1/auth/login", server.authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/login/complete", server.authHandler.HandleLoginComplete)
	mux.HandleFunc("/api/v1/auth/refresh", server.authHandler.HandleRefresh)

	mux.HandleFunc("/api/v1/auth/logout", middleware.WithAuth(server.authHandler.HandleLogout, authClient))
	mux.HandleFunc("/api/v1/auth/2fa/enable", middleware.WithAuth(server.authHandler.HandleEnable2FA, authClient))
	mux.HandleFunc("/api/v1/auth/2fa/enable/complete", middleware.WithAuth(server.authHandler.HandleEnable2FAComplete, authClient))
	mux.HandleFunc("/api/v1/auth/2fa/disable", middleware.WithAuth(server.authHandler.HandleDisable2FA, authClient))
	mux.HandleFunc("/api/v1/auth/2fa/disable/complete", middleware.WithAuth(server.authHandler.HandleDisable2FAComplete, authClient))
	mux.HandleFunc("/api/v1/auth/email/change", middleware.WithAuth(server.authHandler.HandleChangeEmail, authClient))
	mux.HandleFunc("/api/v1/auth/email/change/complete", middleware.WithAuth(server.authHandler.HandleChangeEmailComplete, authClient))
	mux.HandleFunc("/api/v1/auth/password/change", middleware.WithAuth(server.authHandler.HandleChangePassword, authClient))
	mux.HandleFunc("/api/v1/auth/password/change/complete", middleware.WithAuth(server.authHandler.HandleChangePasswordComplete, authClient))
	mux.HandleFunc("/api/v1/auth/meta/change", middleware.WithAuth(server.authHandler.HandleChangeMeta, authClient))

	mux.HandleFunc("/api/v1/files", middleware.WithAuth(server.fileHandler.HandleFiles, authClient))
	mux.HandleFunc("/api/v1/files/", middleware.WithAuth(server.fileHandler.HandleFileDetail, authClient))
	mux.HandleFunc("/api/v1/files/upload", middleware.WithAuth(server.fileHandler.HandleInitiateUpload, authClient))
	mux.HandleFunc("/api/v1/files/upload/complete", middleware.WithAuth(server.fileHandler.HandleCompleteUpload, authClient))
	mux.HandleFunc("/api/v1/files/download/", middleware.WithAuth(server.fileHandler.HandleDownloadLink, authClient))
	mux.HandleFunc("/api/v1/files/trash/", middleware.WithAuth(server.fileHandler.HandleTrashFile, authClient))
	mux.HandleFunc("/api/v1/files/restore/", middleware.WithAuth(server.fileHandler.HandleRestoreFile, authClient))

	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port),
		Handler:      middleware.Metrics(middleware.CORS(mux)),
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	handler.JSONResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/docs/" || r.URL.Path == "/docs/index.html" {
		http.ServeFile(w, r, "docs/swagger/index.html")
		return
	}
	http.StripPrefix("/docs/", http.FileServer(http.Dir("docs/swagger"))).ServeHTTP(w, r)
}

func (s *Server) handleSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "docs/swagger/gateway-openapi.yaml")
}
