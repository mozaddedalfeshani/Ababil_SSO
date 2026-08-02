package totp

import "testing"

func TestGenerateRecoveryCodesUniqueAndHashed(t *testing.T) {
	plaintext, hashes, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(plaintext) != recoveryCodeCount || len(hashes) != recoveryCodeCount {
		t.Fatalf("expected %d codes, got %d plaintext / %d hashes", recoveryCodeCount, len(plaintext), len(hashes))
	}

	seen := map[string]bool{}
	for i, code := range plaintext {
		if seen[code] {
			t.Fatalf("duplicate recovery code generated: %s", code)
		}
		seen[code] = true

		if len(code) != 9 || code[4] != '-' {
			t.Fatalf("unexpected code format: %q", code)
		}
		if hashes[i] == code {
			t.Fatal("hash must not equal plaintext")
		}
	}
}
