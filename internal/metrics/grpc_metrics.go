package metrics

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
)

var (
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	grpcRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_requests_in_flight",
			Help: "Number of gRPC requests currently being processed",
		},
		[]string{"service"},
	)

	authOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_operations_total",
			Help: "Total number of auth operations",
		},
		[]string{"operation", "status"},
	)

	fileOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_operations_total",
			Help: "Total number of file operations",
		},
		[]string{"operation", "status"},
	)

	metadataOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "metadata_operations_total",
			Help: "Total number of metadata operations",
		},
		[]string{"operation", "status"},
	)

	mailOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mail_operations_total",
			Help: "Total number of metadata operations",
		},
		[]string{"operation", "status"},
	)
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		service, method := parseFullMethod(info.FullMethod)

		grpcRequestsInFlight.WithLabelValues(service).Inc()
		defer grpcRequestsInFlight.WithLabelValues(service).Dec()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}

		grpcRequestsTotal.WithLabelValues(service, method, status).Inc()
		grpcRequestDuration.WithLabelValues(service, method).Observe(duration)

		return resp, err
	}
}

func parseFullMethod(fullMethod string) (service, method string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	parts := strings.Split(fullMethod, "/")
	if len(parts) == 2 {
		serviceParts := strings.Split(parts[0], ".")
		service = serviceParts[len(serviceParts)-1]
		method = parts[1]
	}
	return
}

func RecordAuthOperation(operation, status string) {
	authOperationsTotal.WithLabelValues(operation, status).Inc()
}

func RecordFileOperation(operation, status string) {
	fileOperationsTotal.WithLabelValues(operation, status).Inc()
}

func RecordMetadataOperation(operation, status string) {
	metadataOperationsTotal.WithLabelValues(operation, status).Inc()
}

func RecordMailOperation(operation, status string) {
	mailOperationsTotal.WithLabelValues(operation, status).Inc()
}
