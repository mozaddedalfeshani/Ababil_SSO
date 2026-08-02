package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets the baseline header set for every response. CSP
// itself (with its per-request nonce) is wired in Phase 5 alongside the
// UI; this covers the headers that don't depend on page content.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Next()
	}
}
