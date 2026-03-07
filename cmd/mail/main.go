package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/mail"
	"google.golang.org/grpc"
)

func main() {
	config := configs.LoadConfig()

	mailSvc, err := mail.NewMailServiceWithDialer(config)
	if err != nil {
		log.Fatalf("Failed to create mail service: %v", err)
	}

	grpcServer := grpc.NewServer()
	mailServer := mail.NewServer(mailSvc)
	api.RegisterMailServiceServer(grpcServer, mailServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "50054"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down mail service...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Mail service starting on port %s", "50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
