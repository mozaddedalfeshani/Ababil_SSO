package oauth

import "errors"

// ClientError vs. redirect-back errors: RFC 6749 §4.1.2.1. A ClientError
// means the request itself is unsafe to redirect (invalid client_id or
// unregistered redirect_uri) and must be shown as a page, never a
// redirect — otherwise the IdP becomes an open redirector. Every other
// error redirects back to the client with `error`/`state`/`iss`.
type ClientError struct {
	msg string
}

func (e *ClientError) Error() string { return e.msg }

func NewClientError(msg string) error { return &ClientError{msg: msg} }

func IsClientError(err error) bool {
	_, ok := err.(*ClientError)
	return ok
}

// RedirectError carries an OAuth error code that belongs in the
// redirect-back query string, not a rendered page.
type RedirectError struct {
	Code        string
	Description string
}

func (e *RedirectError) Error() string { return e.Code + ": " + e.Description }

func NewRedirectError(code, description string) error {
	return &RedirectError{Code: code, Description: description}
}

var (
	ErrConsentRequired    = errors.New("consent required")
	ErrLoginRequired      = errors.New("login required")
	ErrInvalidGrant       = errors.New("invalid_grant")
	ErrInvalidClient      = errors.New("invalid_client")
	ErrUnauthorizedClient = errors.New("unauthorized_client")
	ErrInvalidScope       = errors.New("invalid_scope")
)
