package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/metadata"
	"github.com/Sene4ka/cloud_storage/internal/metrics"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	config := configs.LoadConfig()

	_, port, err := net.SplitHostPort(config.Services.MetadataAddr)
	if err != nil {
		log.Fatalf("Invalid MetadataAddr format: %s (expected host:port): %v", config.Services.MetadataAddr, err)
	}

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
	metadataSvc := metadata.NewMetadataService(fileRepo)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()))
	metadataServer := metadata.NewServer(metadataSvc)
	api.RegisterMetadataServiceServer(grpcServer, metadataServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("Metrics server starting on port %s", config.Metrics.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", config.Metrics.Port), nil); err != nil {
			log.Fatalf("Failed to serve metrics: %v", err)
		}
	}()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down metadata service...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Metadata service starting on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
