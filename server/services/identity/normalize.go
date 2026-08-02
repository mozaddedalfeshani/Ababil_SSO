package identity

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NormalizeEmail applies NFKC + lowercase before every read/write so
// visually-identical addresses (different Unicode forms, mixed case)
// can't be used to register a second account or bypass a uniqueness
// check the database only enforces on lower(email).
func NormalizeEmail(email string) string {
	return strings.ToLower(norm.NFKC.String(strings.TrimSpace(email)))
}
