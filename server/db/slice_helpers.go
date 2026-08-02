package db

// nonNilStrings guards every NOT NULL text[] column in this schema:
// pgx encodes a nil []string as SQL NULL, not empty-array, and several
// columns here are NOT NULL DEFAULT '{}' with no NULL fallback.
func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
