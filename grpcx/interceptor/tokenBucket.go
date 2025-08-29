package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

type TokenBucketBuilder struct {
	mu       sync.Mutex
	bucket   chan struct{}
	interval time.Duration
	cancel   context.CancelFunc
	enabled  atomic.Bool
}

func NewTokenBucketBuilder(interval time.Duration, size int64) *TokenBucketBuilder {
	if size <= 0 {
		panic("token bucket size must be > 0")
	}

	t := &TokenBucketBuilder{
		bucket:   make(chan struct{}, size),
		interval: interval,
	}

	for i := 0; i < cap(t.bucket); i++ {
		t.bucket <- struct{}{}
	}

	t.Enable()
	return t
}

func (t *TokenBucketBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !t.enabled.Load() {
			return handler(ctx, req)
		}
		select {
		case <-t.bucket:
			return handler(ctx, req)
		default:
			return nil, status.Error(codes.ResourceExhausted, "rate limited")
		}
	}
}

func (t *TokenBucketBuilder) start(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case t.bucket <- struct{}{}:
			default:
			}
		}
	}
}

func (t *TokenBucketBuilder) Disable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.enabled.Load() {
		return
	}
	t.enabled.Store(false)
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

func (t *TokenBucketBuilder) Enable() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.enabled.Load() {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.enabled.Store(true)
	go t.start(ctx)
}
