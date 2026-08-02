package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestSealOpenRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte("totp secret material")
	aad := []byte("user-id-123")

	ct, err := Seal(key, plaintext, aad)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	pt, err := Open(key, ct, aad)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Fatalf("plaintext mismatch: got %q want %q", pt, plaintext)
	}
}

func TestOpenRejectsWrongAAD(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	ct, err := Seal(key, []byte("secret"), []byte("aad-a"))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	if _, err := Open(key, ct, []byte("aad-b")); err == nil {
		t.Fatal("expected AAD mismatch to fail decryption")
	}
}

func TestHashTokenDeterministicAndDistinct(t *testing.T) {
	h1 := HashToken("token-a")
	h2 := HashToken("token-a")
	h3 := HashToken("token-b")

	if h1 != h2 {
		t.Fatal("expected same input to hash identically")
	}
	if h1 == h3 {
		t.Fatal("expected different input to hash differently")
	}
}
