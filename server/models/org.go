package models

import "time"

type Organization struct {
	ID          string
	Name        string
	Slug        string
	OwnerUserID string
	CreatedAt   time.Time
}

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

type OrganizationMember struct {
	OrgID     string
	UserID    string
	Role      OrgRole
	CreatedAt time.Time
}
