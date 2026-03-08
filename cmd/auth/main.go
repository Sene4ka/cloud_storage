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
	"github.com/Sene4ka/cloud_storage/internal/auth"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	defer redisClient.Close()

	userRepo := repositories.NewUserRepository(dbpool)

	mailCC, err := grpc.NewClient(config.Services.MailAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("mail conn: %w", err)
	}

	mailConn := grpc.ClientConnInterface(mailCC)
	mailClient := api.NewMailServiceClient(mailConn)

	authService := auth.NewAuthServiceRedis(userRepo, redisClient, mailClient, config)

	grpcServer := grpc.NewServer()
	authServer := auth.NewServer(authService)
	api.RegisterAuthServiceServer(grpcServer, authServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "50051"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down auth service...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Auth service starting on port %s", "50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
