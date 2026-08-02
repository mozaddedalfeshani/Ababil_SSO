package handlers

import (
	"html"
	"net/http"
	"net/url"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthLogoutHandler struct {
	OAuth      *oauth.Service
	AppBaseURL string
}

// Logout implements RP-initiated logout's *front-channel entry point*
// only: it validates the request (see oauth.ResolveLogoutClient) and
// hands off to Next's confirmation screen. The actual session
// revocation happens on explicit user confirmation via
// POST /api/oauth/logout/confirm — a bare GET must never be able to
// end a session, or this becomes CSRF-able logout (see
// docs/architecture.md "Logout hardening").
func (h *OAuthLogoutHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	q := c.Request.URL.Query()

	idTokenHint := q.Get("id_token_hint")
	postLogout := q.Get("post_logout_redirect_uri")

	if _, err := h.OAuth.ResolveLogoutClient(ctx, idTokenHint, postLogout); err != nil {
		body := `<!doctype html><html><body><h1>Logout request error</h1><p>` +
			html.EscapeString(err.Error()) + `</p></body></html>`
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(body))
		return
	}

	dest, err := url.Parse(h.AppBaseURL + "/logout")
	if err != nil {
		respondInternalError(c, err)
		return
	}
	dq := dest.Query()
	dq.Set("id_token_hint", idTokenHint)
	dq.Set("post_logout_redirect_uri", postLogout)
	dq.Set("state", q.Get("state"))
	dest.RawQuery = dq.Encode()

	c.Redirect(http.StatusFound, dest.String())
}
