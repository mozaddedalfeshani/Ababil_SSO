package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/models"
	"ababilx-sso/services/orgs"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	Orgs *orgs.Service
}

type createClientRequest struct {
	Name           string             `json:"name" binding:"required"`
	ClientType     models.ClientType  `json:"client_type" binding:"required"`
	RedirectURIs   []string           `json:"redirect_uris" binding:"required,min=1"`
	PostLogoutURIs []string           `json:"post_logout_redirect_uris"`
	AllowedScopes  []string           `json:"allowed_scopes" binding:"required,min=1"`
	SubjectType    models.SubjectType `json:"subject_type"`
	RequireConsent *bool              `json:"require_consent"`
}

// Create requires org admin — creating a client is a credential-issuing
// action, not something a plain member should be able to do
// unsupervised.
func (h *ClientHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	userID := middleware.UserID(c)

	if err := h.Orgs.RequireRole(ctx, orgID, userID, models.OrgRoleAdmin); err != nil {
		respondForbiddenOrError(c, err)
		return
	}

	var req createClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if req.ClientType != models.ClientTypePublic && req.ClientType != models.ClientTypeConfidential {
		respondError(c, http.StatusBadRequest, "invalid_request", "client_type must be public or confidential")
		return
	}

	requireConsent := true
	if req.RequireConsent != nil {
		requireConsent = *req.RequireConsent
	}

	grantTypes := []string{"authorization_code", "refresh_token"}
	if req.ClientType == models.ClientTypeConfidential {
		grantTypes = append(grantTypes, "client_credentials")
	}

	client, secret, err := h.Orgs.CreateClient(ctx, orgs.CreateClientParams{
		OrgID:          orgID,
		Name:           req.Name,
		ClientType:     req.ClientType,
		RedirectURIs:   req.RedirectURIs,
		PostLogoutURIs: req.PostLogoutURIs,
		GrantTypes:     grantTypes,
		AllowedScopes:  req.AllowedScopes,
		SubjectType:    req.SubjectType,
		RequireConsent: requireConsent,
		CreatedBy:      userID,
	})
	if err != nil {
		respondInternalError(c, err)
		return
	}

	resp := gin.H{"client": clientView(client)}
	if secret != "" {
		// Shown exactly once — see services/orgs.CreateClient.
		resp["client_secret"] = secret
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *ClientHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	if err := h.Orgs.RequireRole(ctx, orgID, middleware.UserID(c), models.OrgRoleMember); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	clients, err := h.Orgs.Clients.ListForOrg(ctx, orgID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"clients": clientViews(clients)})
}

func (h *ClientHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	client, err := h.Orgs.Clients.ByID(ctx, c.Param("clientId"))
	if err != nil {
		respondError(c, http.StatusNotFound, "not_found", "client not found")
		return
	}
	if err := h.Orgs.RequireRole(ctx, client.OrgID, middleware.UserID(c), models.OrgRoleMember); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	c.JSON(http.StatusOK, clientView(client))
}

func (h *ClientHandler) RotateSecret(c *gin.Context) {
	ctx := c.Request.Context()
	clientID := c.Param("clientId")

	client, err := h.Orgs.Clients.ByID(ctx, clientID)
	if err != nil {
		respondError(c, http.StatusNotFound, "not_found", "client not found")
		return
	}
	if err := h.Orgs.RequireRole(ctx, client.OrgID, middleware.UserID(c), models.OrgRoleAdmin); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	if !client.IsConfidential() {
		respondError(c, http.StatusConflict, "not_confidential", "public clients have no secret to rotate")
		return
	}

	secret, err := h.Orgs.RotateSecret(ctx, clientID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_secret": secret})
}

func (h *ClientHandler) Disable(c *gin.Context) {
	ctx := c.Request.Context()
	clientID := c.Param("clientId")

	client, err := h.Orgs.Clients.ByID(ctx, clientID)
	if err != nil {
		respondError(c, http.StatusNotFound, "not_found", "client not found")
		return
	}
	if err := h.Orgs.RequireRole(ctx, client.OrgID, middleware.UserID(c), models.OrgRoleAdmin); err != nil {
		respondForbiddenOrError(c, err)
		return
	}
	if err := h.Orgs.Clients.Disable(ctx, clientID); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client disabled"})
}
