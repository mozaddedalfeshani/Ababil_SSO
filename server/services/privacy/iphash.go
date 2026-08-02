// Package privacy hashes IP addresses under a salt that rotates daily,
// so audit rows never store a raw IP but can still be correlated
// within a single day (e.g. for lockout/abuse investigation) without
// remaining linkable indefinitely.
package privacy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"ababilx-sso/cache"
	"ababilx-sso/services/crypto"

	"github.com/redis/go-redis/v9"
)

type IPHasher struct {
	redis *redis.Client
}

func NewIPHasher(redis *redis.Client) *IPHasher { return &IPHasher{redis: redis} }

// Hash returns HMAC-SHA256(ip, daily_salt). The salt is generated on
// first use each day and cached in Redis with a short-lived TTL past
// midnight, so it naturally expires and rotates without a cron job.
func (h *IPHasher) Hash(ctx context.Context, ip string) (string, error) {
	salt, err := h.dailySalt(ctx)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (h *IPHasher) dailySalt(ctx context.Context) ([]byte, error) {
	key := cache.IPSaltKey(time.Now().UTC().Format("20060102"))

	existing, err := h.redis.Get(ctx, key).Result()
	if err == nil {
		return []byte(existing), nil
	}
	if err != redis.Nil {
		return nil, err
	}

	salt, err := crypto.RandomToken(32)
	if err != nil {
		return nil, err
	}
	// 26h TTL: comfortably past midnight even with clock skew, still
	// bounded so the salt doesn't persist indefinitely. SetNX so a
	// concurrent first-use of the day can't race two different salts
	// into existence — the loser re-reads what the winner set.
	ok, err := h.redis.SetNX(ctx, key, salt, 26*time.Hour).Result()
	if err != nil {
		return nil, err
	}
	if ok {
		return []byte(salt), nil
	}

	winner, err := h.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(winner), nil
}
