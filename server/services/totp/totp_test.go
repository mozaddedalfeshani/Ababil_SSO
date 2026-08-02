package totp

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
	baseTotp "github.com/pquerna/otp/totp"
)

func genCode(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	code, err := baseTotp.GenerateCodeCustom(secret, at, baseTotp.ValidateOpts{
		Period:    stepSeconds,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	return code
}

func TestVerifyAcceptsCurrentCode(t *testing.T) {
	secret, err := GenerateSecret("Test Issuer", "user@example.com")
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}

	now := time.Now()
	code := genCode(t, secret.Base32, now)

	step, ok, err := Verify(code, secret.Base32, 0, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected current code to verify")
	}
	if step != now.Unix()/stepSeconds {
		t.Fatalf("unexpected step: %d", step)
	}
}

func TestVerifyRejectsAlreadyConsumedStep(t *testing.T) {
	secret, err := GenerateSecret("Test Issuer", "user@example.com")
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}

	now := time.Now()
	code := genCode(t, secret.Base32, now)
	currentStep := now.Unix() / stepSeconds

	// Simulate the step already being recorded as consumed.
	_, ok, err := Verify(code, secret.Base32, currentStep, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("expected replay of an already-consumed step to be rejected")
	}
}

func TestVerifyRejectsWrongCode(t *testing.T) {
	secret, err := GenerateSecret("Test Issuer", "user@example.com")
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}

	_, ok, err := Verify("000000", secret.Base32, 0, time.Now())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("expected wrong code to fail")
	}
}

func TestVerifyRejectsAdjacentWindow(t *testing.T) {
	secret, err := GenerateSecret("Test Issuer", "user@example.com")
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}

	now := time.Now()
	// Code from the previous 30s step must NOT verify — Skew is
	// deliberately 0 (see totp.go), tightening the replay window.
	previousStepTime := now.Add(-stepSeconds * time.Second)
	code := genCode(t, secret.Base32, previousStepTime)

	_, ok, err := Verify(code, secret.Base32, 0, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("expected adjacent-window code to be rejected (Skew=0)")
	}
}
