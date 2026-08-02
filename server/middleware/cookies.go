package middleware

// CookiePolicy names cookies and carries the attributes every auth
// cookie must share. __Host- names require Secure, Path=/, and no
// Domain attribute — browsers enforce this by silently refusing to
// store a __Host- cookie that lacks Secure, which is exactly what
// happens over plain HTTP local dev. So the prefix is only used when
// BOTH Domain is unset AND Secure is true; anything else (subdomain
// deployments, or plain-HTTP dev) falls back to plain names.
type CookiePolicy struct {
	Domain string // "" => eligible for __Host- cookies (recommended)
	Secure bool   // false only for http://localhost dev
}

func (p CookiePolicy) usesHostPrefix() bool { return p.Domain == "" && p.Secure }

func (p CookiePolicy) SessionCookieName() string {
	if p.usesHostPrefix() {
		return "__Host-sso_session"
	}
	return "sso_session"
}

func (p CookiePolicy) CSRFCookieName() string {
	if p.usesHostPrefix() {
		return "__Host-sso_csrf"
	}
	return "sso_csrf"
}

const CSRFHeaderName = "X-CSRF-Token"
