package handlers

import (
	"net/http"

	"ababilx-sso/middleware"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) Me(c *gin.Context) {
	user, err := h.Identity.Users.ByID(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerified(),
		"totp_enabled":   user.TOTPEnabled(),
		"created_at":     user.CreatedAt,
		"last_login_at":  user.LastLoginAt,
	})
}
