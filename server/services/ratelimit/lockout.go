package ratelimit

import (
	"context"
	"time"

	"ababilx-sso/cache"

	"github.com/redis/go-redis/v9"
)

// Lockout tracks consecutive login failures per account (keyed by a
// hash of the email, never the raw address) and grows the lockout
// window with each additional failure, so a slow distributed guesser
// against one account is throttled harder the longer it persists.
type Lockout struct {
	redis *redis.Client
}

func NewLockout(redis *redis.Client) *Lockout { return &Lockout{redis: redis} }

// thresholds[i] is the lockout duration once failures reach index i+1.
var thresholds = []struct {
	failures int
	lockout  time.Duration
}{
	{5, 30 * time.Second},
	{10, 5 * time.Minute},
	{15, 30 * time.Minute},
	{20, 2 * time.Hour},
}

// Locked reports whether the account identified by emailHash is
// currently locked out. Fails closed: a Redis error is treated as
// locked, matching the fail-closed policy for auth endpoints.
func (l *Lockout) Locked(ctx context.Context, emailHash string) (bool, error) {
	ttl, err := l.redis.TTL(ctx, cache.LoginFailKey(emailHash)+":locked").Result()
	if err != nil {
		return true, err
	}
	return ttl > 0, nil
}

// RecordFailure increments the failure counter and, once a threshold
// is crossed, sets the lockout key with the matching TTL.
func (l *Lockout) RecordFailure(ctx context.Context, emailHash string) error {
	key := cache.LoginFailKey(emailHash)
	count, err := l.redis.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		l.redis.Expire(ctx, key, 24*time.Hour)
	}

	lockDuration := time.Duration(0)
	for _, t := range thresholds {
		if int(count) >= t.failures {
			lockDuration = t.lockout
		}
	}
	if lockDuration > 0 {
		if err := l.redis.Set(ctx, key+":locked", "1", lockDuration).Err(); err != nil {
			return err
		}
	}
	return nil
}

// Clear resets both the failure count and any active lockout —
// called on successful login.
func (l *Lockout) Clear(ctx context.Context, emailHash string) error {
	key := cache.LoginFailKey(emailHash)
	return l.redis.Del(ctx, key, key+":locked").Err()
}
