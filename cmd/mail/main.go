package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/mail"
	"github.com/Sene4ka/cloud_storage/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	config := configs.LoadConfig()

	_, port, err := net.SplitHostPort(config.Services.MailAddr)
	if err != nil {
		log.Fatalf("Invalid MailAddr format: %s (expected host:port): %v", config.Services.MailAddr, err)
	}

	mailSvc, err := mail.NewMailServiceWithDialer(config)
	if err != nil {
		log.Fatalf("Failed to create mail service: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()))
	mailServer := mail.NewServer(mailSvc)
	api.RegisterMailServiceServer(grpcServer, mailServer)

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
		log.Println("Shutting down mail service...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Mail service starting on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
