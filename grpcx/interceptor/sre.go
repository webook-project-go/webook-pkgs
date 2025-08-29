package interceptor

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kratos/aegis/circuitbreaker"
)

var ErrRequestRejected = errors.New("request rejected by SRE breaker")

type SreInterceptor struct {
	breaker circuitbreaker.CircuitBreaker
}

func NewSreInterceptor(cb circuitbreaker.CircuitBreaker) *SreInterceptor {
	return &SreInterceptor{
		breaker: cb,
	}
}

func (s *SreInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 检查是否允许请求
		if s.breaker != nil {
			if err := s.breaker.Allow(); err != nil {
				if s.breaker != nil {
					s.breaker.MarkFailed()
				}
				return nil, status.Error(codes.ResourceExhausted, ErrRequestRejected.Error())
			}
		}
		resp, err = handler(ctx, req)
		if s.breaker != nil {
			if err != nil {
				s.breaker.MarkFailed()
			} else {
				s.breaker.MarkSuccess()
			}
		}

		return resp, err
	}
}
