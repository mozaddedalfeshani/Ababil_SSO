package handlers

import "ababilx-sso/models"

// clientView is the client-management API's shape — deliberately not
// the raw *models.OAuthClient, which carries ClientSecretHMAC. That
// field has no reason to leave the server even hashed (same class of
// issue as leaking a session's token_hash): it is verification
// material, not something any caller needs to read back.
func clientView(c *models.OAuthClient) map[string]any {
	return map[string]any{
		"id":                         c.ID,
		"org_id":                     c.OrgID,
		"client_id":                  c.ClientID,
		"name":                       c.Name,
		"logo_url":                   c.LogoURL,
		"client_type":                c.ClientType,
		"token_endpoint_auth_method": c.TokenEndpointAuthMethod,
		"redirect_uris":              c.RedirectURIs,
		"post_logout_redirect_uris":  c.PostLogoutRedirectURIs,
		"grant_types":                c.GrantTypes,
		"allowed_scopes":             c.AllowedScopes,
		"subject_type":               c.SubjectType,
		"require_consent":            c.RequireConsent,
		"disabled":                   c.Disabled(),
		"created_at":                 c.CreatedAt,
	}
}

func clientViews(clients []*models.OAuthClient) []map[string]any {
	out := make([]map[string]any, 0, len(clients))
	for _, c := range clients {
		out = append(out, clientView(c))
	}
	return out
}
