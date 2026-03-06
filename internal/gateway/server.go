package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/gateway/handler"
	"github.com/Sene4ka/cloud_storage/internal/gateway/middleware"
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
	mux.HandleFunc("/api/v1/auth/register", server.authHandler.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", server.authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", server.authHandler.HandleRefresh)
	mux.HandleFunc("/api/v1/auth/logout", middleware.WithAuth(server.authHandler.HandleLogout, authClient))
	mux.HandleFunc("/api/v1/files", middleware.WithAuth(server.fileHandler.HandleFiles, authClient))
	mux.HandleFunc("/api/v1/files/", middleware.WithAuth(server.fileHandler.HandleFileDetail, authClient))
	mux.HandleFunc("/api/v1/files/upload", middleware.WithAuth(server.fileHandler.HandleInitiateUpload, authClient))
	mux.HandleFunc("/api/v1/files/upload/complete", middleware.WithAuth(server.fileHandler.HandleCompleteUpload, authClient))
	mux.HandleFunc("/api/v1/files/download/", middleware.WithAuth(server.fileHandler.HandleDownloadLink, authClient))
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port),
		Handler:      middleware.CORS(mux),
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
