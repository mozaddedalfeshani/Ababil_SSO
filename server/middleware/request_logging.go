package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogging logs method, path, status, and latency only — no IP,
// no query string, no body. Raw client IPs are privacy-sensitive and
// are never written to process logs; the audit trail stores only an
// HMAC of the IP, and only for security-relevant events (see
// services/audit).
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		slog.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}
