package handlers

import (
	"errors"
	"net/http"
	"strings"

	"ababilx-sso/middleware"
	"ababilx-sso/models"
	"ababilx-sso/services/orgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrgHandler struct {
	Orgs *orgs.Service
}

type createOrgRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *OrgHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	var req createOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	slug := slugify(req.Name) + "-" + uuid.NewString()[:8]
	org, err := h.Orgs.Orgs.Create(ctx, req.Name, slug, userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, org)
}

func (h *OrgHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.Orgs.Orgs.ListForUser(ctx, middleware.UserID(c))
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"organizations": list})
}

func (h *OrgHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	if err := h.Orgs.RequireRole(ctx, orgID, middleware.UserID(c), models.OrgRoleMember); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	org, err := h.Orgs.Orgs.ByID(ctx, orgID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, org)
}

type addMemberRequest struct {
	UserID string         `json:"user_id" binding:"required"`
	Role   models.OrgRole `json:"role" binding:"required"`
}

func (h *OrgHandler) AddMember(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	if err := h.Orgs.RequireRole(ctx, orgID, middleware.UserID(c), models.OrgRoleAdmin); err != nil {
		respondForbiddenOrError(c, err)
		return
	}

	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Role != models.OrgRoleAdmin && req.Role != models.OrgRoleMember) {
		respondError(c, http.StatusBadRequest, "invalid_request", "user_id and role (admin|member) are required")
		return
	}

	if err := h.Orgs.Orgs.AddMember(ctx, orgID, req.UserID, req.Role); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *OrgHandler) ListMembers(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	if err := h.Orgs.RequireRole(ctx, orgID, middleware.UserID(c), models.OrgRoleMember); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	members, err := h.Orgs.Orgs.ListMembers(ctx, orgID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *OrgHandler) RemoveMember(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.Orgs.RequireRole(ctx, orgID, middleware.UserID(c), models.OrgRoleAdmin); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	if err := h.Orgs.Orgs.RemoveMember(ctx, orgID, targetUserID); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

func respondForbiddenOrError(c *gin.Context, err error) {
	if errors.Is(err, orgs.ErrForbidden) {
		respondError(c, http.StatusForbidden, "forbidden", "insufficient permissions for this organization")
		return
	}
	respondInternalError(c, err)
}

func slugify(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	s := b.String()
	if s == "" {
		s = "org"
	}
	return s
}
