package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/metadata"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	config := configs.LoadConfig()
	dbpool, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
		config.Database.SSLMode,
	))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	fileRepo := repositories.NewFileRepository(dbpool)
	grpcServer := grpc.NewServer()
	metadataServer := metadata.NewServer(fileRepo)
	api.RegisterMetadataServiceServer(grpcServer, metadataServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "50052"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down metadata service...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Metadata service starting on port %s", "50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
