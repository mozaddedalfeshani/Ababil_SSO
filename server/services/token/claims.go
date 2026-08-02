// Package token signs and assembles OAuth access tokens and OIDC ID
// tokens. It never touches the database — callers pass in everything
// needed (subject, scope, key) so this stays a pure function layer.
package token

// AccessTokenClaims follows RFC 9068 (JWT access tokens). `sub` is
// always the pairwise subject for the audience client — see
// services/token/pairwise.go — never a value shared across clients.
type AccessTokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	ClientID  string `json:"client_id"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Scope     string `json:"scope"`
	JTI       string `json:"jti"`
}

// IDTokenClaims follows OpenID Connect Core. Email claims are included
// only when the `email` scope was granted — callers decide that
// before constructing this struct.
type IDTokenClaims struct {
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	Audience      string   `json:"aud"`
	ExpiresAt     int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
	AuthTime      int64    `json:"auth_time"`
	Nonce         string   `json:"nonce,omitempty"`
	AMR           []string `json:"amr,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified *bool    `json:"email_verified,omitempty"`
}
