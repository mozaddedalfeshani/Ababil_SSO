// Package ratelimit implements a Redis fixed-window counter. Auth
// endpoints (login, register, reset, TOTP verify, token, introspect)
// call Allow with FailMode=Closed: if Redis is unreachable, the
// request is rejected. An attacker who can knock over Redis must not
// thereby unlock unlimited credential stuffing. Read-only/dashboard
// traffic uses FailMode=Open, where availability wins.
package ratelimit

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/cache"

	"github.com/redis/go-redis/v9"
)

type FailMode int

const (
	FailClosed FailMode = iota
	FailOpen
)

var ErrRateLimited = errors.New("rate limit exceeded")

type Limiter struct {
	redis *redis.Client
}

func NewLimiter(redis *redis.Client) *Limiter { return &Limiter{redis: redis} }

// Allow increments the counter for (bucket, key) and errors with
// ErrRateLimited if it exceeds max within window. bucket names a
// specific policy (e.g. "login_ip", "oauth_token") so different call
// sites never share a counter by accident.
func (l *Limiter) Allow(ctx context.Context, bucket, key string, max int, window time.Duration, failMode FailMode) error {
	redisKey := cache.RateLimitKey(bucket, key)

	count, err := l.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		if failMode == FailOpen {
			return nil
		}
		return err
	}
	if count == 1 {
		l.redis.Expire(ctx, redisKey, window)
	}

	if int(count) > max {
		return ErrRateLimited
	}
	return nil
}

// Reset clears a counter — used after a successful login to zero the
// per-account failure count rather than waiting out the window.
func (l *Limiter) Reset(ctx context.Context, bucket, key string) error {
	return l.redis.Del(ctx, cache.RateLimitKey(bucket, key)).Err()
}
