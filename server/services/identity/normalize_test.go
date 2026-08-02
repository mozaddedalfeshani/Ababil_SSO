package identity

import "testing"

func TestNormalizeEmail(t *testing.T) {
	cases := map[string]string{
		"User@Example.com":  "user@example.com",
		"  a@b.com  ":       "a@b.com",
		"ALREADY@lower.com": "already@lower.com",
	}
	for in, want := range cases {
		if got := NormalizeEmail(in); got != want {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeEmailStableUnderRepeatedApplication(t *testing.T) {
	in := "Mixed.Case+Tag@Example.COM"
	once := NormalizeEmail(in)
	twice := NormalizeEmail(once)
	if once != twice {
		t.Fatalf("normalization not idempotent: %q vs %q", once, twice)
	}
}
