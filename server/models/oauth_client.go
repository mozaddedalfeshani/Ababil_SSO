package models

import "time"

type ClientType string

const (
	ClientTypePublic       ClientType = "public"
	ClientTypeConfidential ClientType = "confidential"
)

type TokenEndpointAuthMethod string

const (
	AuthMethodClientSecretBasic TokenEndpointAuthMethod = "client_secret_basic"
	AuthMethodClientSecretPost  TokenEndpointAuthMethod = "client_secret_post"
	AuthMethodNone              TokenEndpointAuthMethod = "none"
)

type SubjectType string

const (
	SubjectTypePairwise SubjectType = "pairwise"
	SubjectTypePublic   SubjectType = "public"
)

type OAuthClient struct {
	ID                      string
	OrgID                   string
	ClientID                string
	ClientSecretHMAC        *string
	Name                    string
	LogoURL                 *string
	ClientType              ClientType
	TokenEndpointAuthMethod TokenEndpointAuthMethod
	RedirectURIs            []string
	PostLogoutRedirectURIs  []string
	GrantTypes              []string
	AllowedScopes           []string
	SubjectType             SubjectType
	SectorIdentifier        *string
	RequireConsent          bool
	CreatedBy               string
	DisabledAt              *time.Time
	CreatedAt               time.Time
}

func (c *OAuthClient) IsConfidential() bool { return c.ClientType == ClientTypeConfidential }
func (c *OAuthClient) Disabled() bool       { return c.DisabledAt != nil }

func (c *OAuthClient) SupportsGrant(grant string) bool {
	for _, g := range c.GrantTypes {
		if g == grant {
			return true
		}
	}
	return false
}

func (c *OAuthClient) AllowsRedirectURI(uri string) bool {
	for _, u := range c.RedirectURIs {
		if u == uri {
			return true
		}
	}
	return false
}

func (c *OAuthClient) AllowsPostLogoutRedirectURI(uri string) bool {
	for _, u := range c.PostLogoutRedirectURIs {
		if u == uri {
			return true
		}
	}
	return false
}

// AllowsScopes reports whether every requested scope is in the
// client's allow-list — used to reject scope escalation at request
// time, before it ever reaches consent.
func (c *OAuthClient) AllowsScopes(requested []string) bool {
	allowed := make(map[string]bool, len(c.AllowedScopes))
	for _, s := range c.AllowedScopes {
		allowed[s] = true
	}
	for _, s := range requested {
		if !allowed[s] {
			return false
		}
	}
	return true
}
