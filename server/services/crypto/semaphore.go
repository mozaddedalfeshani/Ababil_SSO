package crypto

import (
	"context"
	"errors"
)

var ErrHashQueueTimeout = errors.New("password hashing queue timed out")

// HashSemaphore bounds concurrent Argon2id operations. Argon2id at
// 64 MB/hash is memory-hard by design, which means a flood of
// concurrent login attempts exhausts RAM long before it exhausts CPU.
// Rate limiting alone doesn't stop this — an attacker distributed
// across many IPs/accounts still gets through the limiter one request
// at a time while all of them hash concurrently. The semaphore caps
// how many hash operations run at once; anything beyond that queues
// with a timeout instead of piling up unbounded and OOM-killing the
// process.
type HashSemaphore struct {
	slots chan struct{}
}

func NewHashSemaphore(maxConcurrent int) *HashSemaphore {
	return &HashSemaphore{slots: make(chan struct{}, maxConcurrent)}
}

// Acquire blocks until a slot is free or ctx is done. Callers should
// pass a context with a short timeout (a few seconds) — a caller that
// waits forever just becomes the next form of resource exhaustion.
func (s *HashSemaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrHashQueueTimeout
	}
}

func (s *HashSemaphore) Release() {
	<-s.slots
}
