package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError writes the dashboard/account API error envelope:
// {"error":{"code","message"}}. OAuth protocol endpoints (Phase 3) use
// the RFC-shaped {"error","error_description"} envelope instead —
// deliberately different shapes so a client can't confuse which API
// surface it's talking to.
func respondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func respondInternalError(c *gin.Context, err error) {
	slog.Error("internal_error", "path", c.Request.URL.Path, "error", err)
	respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
}
