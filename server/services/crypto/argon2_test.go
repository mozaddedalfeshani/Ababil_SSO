package crypto

import (
	"context"
	"testing"
	"time"
)

func testHasher(t *testing.T) *PasswordHasher {
	t.Helper()
	return NewPasswordHasher(DefaultArgon2Params(), NewHashSemaphore(4))
}

func TestArgon2RoundTrip(t *testing.T) {
	h := testHasher(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	encoded, err := h.Hash(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	ok, err := h.Verify(ctx, "correct horse battery staple", encoded)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
}

func TestArgon2WrongPassword(t *testing.T) {
	h := testHasher(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	encoded, err := h.Hash(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	ok, err := h.Verify(ctx, "wrong password", encoded)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestArgon2DummyHashNeverVerifies(t *testing.T) {
	h := testHasher(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ok, err := h.Verify(ctx, "anything at all", DummyHash)
	if err != nil {
		t.Fatalf("verify against dummy hash should not error: %v", err)
	}
	if ok {
		t.Fatal("dummy hash must never verify")
	}
}

func TestHashSemaphoreBoundsConcurrency(t *testing.T) {
	sem := NewHashSemaphore(1)
	ctx := context.Background()

	if err := sem.Acquire(ctx); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := sem.Acquire(timeoutCtx); err != ErrHashQueueTimeout {
		t.Fatalf("expected queue timeout while slot held, got %v", err)
	}

	sem.Release()
	if err := sem.Acquire(context.Background()); err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
}
