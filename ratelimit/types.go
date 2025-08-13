package ratelimit

import "context"

type RateLimiter interface {
	Limit(ctx context.Context, key string) (bool, error)
}
