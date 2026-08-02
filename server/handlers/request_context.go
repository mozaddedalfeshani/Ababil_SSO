package handlers

import (
	"context"

	"ababilx-sso/services/privacy"

	"github.com/gin-gonic/gin"
)

// clientContext resolves the hashed IP and truncated user agent used
// for session rows and audit events. c.ClientIP() already respects
// the trusted-proxy configuration set in routes.NewEngine.
func clientContext(ctx context.Context, c *gin.Context, hasher *privacy.IPHasher) (ipHash *string, userAgent *string) {
	if hash, err := hasher.Hash(ctx, c.ClientIP()); err == nil {
		ipHash = &hash
	}
	if ua := c.Request.UserAgent(); ua != "" {
		if len(ua) > 256 {
			ua = ua[:256]
		}
		userAgent = &ua
	}
	return ipHash, userAgent
}
