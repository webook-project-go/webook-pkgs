package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/slide_window.lua
var luaRedisSlideWindow string

type RedisSlideWindowRateLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	rate     int
}

func NewRedisSlideWindowRateLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowRateLimiter {
	return &RedisSlideWindowRateLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlideWindowRateLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaRedisSlideWindow, []string{key},
		r.interval.Milliseconds(),
		r.rate,
		time.Now().UnixMilli(),
	).Bool()
}
