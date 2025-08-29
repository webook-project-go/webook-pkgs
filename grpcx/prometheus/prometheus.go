package prometheus

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusBuilder struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewPrometheusBuilder() *PrometheusBuilder {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requests, duration)

	return &PrometheusBuilder{
		requests: requests,
		duration: duration,
	}
}

func (p *PrometheusBuilder) BuildUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)

		statusCode := "OK"
		if err != nil {
			s, _ := status.FromError(err)
			statusCode = s.Code().String()
		}

		p.requests.WithLabelValues(info.FullMethod, statusCode).Inc()
		p.duration.WithLabelValues(info.FullMethod).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
