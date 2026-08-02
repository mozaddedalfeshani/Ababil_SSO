package handlers

import (
	"html"
	"net/http"
	"net/url"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthAuthorizeHandler struct {
	OAuth      *oauth.Service
	AppBaseURL string // Next.js UI base — the consent screen lives here
}

// Authorize implements RFC 6749 §4.1.2.1 error handling precisely:
// invalid client_id or unregistered redirect_uri renders a page (this
// endpoint must never become an open redirector for those cases), and
// every other error redirects back to the RP with `error` — see
// oauth.ClientError vs oauth.RedirectError and
// docs/architecture.md.
func (h *OAuthAuthorizeHandler) Authorize(c *gin.Context) {
	ctx := c.Request.Context()
	q := c.Request.URL.Query()

	id, err := h.OAuth.StartAuthorize(ctx, oauth.StartAuthorizeParams{
		ClientID:            q.Get("client_id"),
		RedirectURI:         q.Get("redirect_uri"),
		ResponseType:        q.Get("response_type"),
		Scope:               q.Get("scope"),
		State:               q.Get("state"),
		Nonce:               q.Get("nonce"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		Prompt:              q.Get("prompt"),
	})

	if oauth.IsClientError(err) {
		// Rendered directly, never redirected: the client_id/redirect_uri
		// themselves are unverified at this point, so redirecting would
		// make this endpoint an open redirector (RFC 6749 §4.1.2.1).
		body := `<!doctype html><html><body><h1>Authorization request error</h1><p>` +
			html.EscapeString(err.Error()) + `</p></body></html>`
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(body))
		return
	}
	if redirErr, ok := err.(*oauth.RedirectError); ok {
		c.Redirect(http.StatusFound, buildErrorRedirect(q.Get("redirect_uri"), redirErr.Code, redirErr.Description, q.Get("state")))
		return
	}
	if err != nil {
		respondInternalError(c, err)
		return
	}

	dest, parseErr := url.Parse(h.AppBaseURL + "/authorize")
	if parseErr != nil {
		respondInternalError(c, parseErr)
		return
	}
	dq := dest.Query()
	dq.Set("req", id)
	dest.RawQuery = dq.Encode()

	c.Redirect(http.StatusFound, dest.String())
}

func buildErrorRedirect(redirectURI, code, description, state string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	q := u.Query()
	q.Set("error", code)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
