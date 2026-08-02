package handlers

import (
	"net/http"

	"ababilx-sso/middleware"

	"github.com/gin-gonic/gin"
)

// RecoveryCodesStatus reports only the remaining count — the codes
// themselves are shown exactly once, at enrollment, and can never be
// re-fetched. This endpoint exists so the account UI can prompt
// "running low, regenerate?" without ever exposing a code again.
func (h *AccountHandler) RecoveryCodesStatus(c *gin.Context) {
	ctx := c.Request.Context()
	remaining, err := h.Identity.RecoveryCodes.CountRemaining(ctx, middleware.UserID(c))
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"remaining": remaining})
}
