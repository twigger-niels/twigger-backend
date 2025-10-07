package entity

import (
	"time"

	"github.com/google/uuid"
)

// Workspace represents a multi-tenant workspace
type Workspace struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkspaceMember represents a user's membership in a workspace
type WorkspaceMember struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"` // 'admin', 'member', 'viewer'
	JoinedAt    time.Time `json:"joined_at"`
}

// WorkspaceRole represents possible roles in a workspace
type WorkspaceRole string

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleViewer WorkspaceRole = "viewer"
)

// IsValid checks if the role is valid
func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}

// IsAdmin returns true if the role is admin
func (wm *WorkspaceMember) IsAdmin() bool {
	return wm.Role == string(RoleAdmin)
}

// IsMember returns true if the role is member or admin
func (wm *WorkspaceMember) IsMember() bool {
	return wm.Role == string(RoleMember) || wm.IsAdmin()
}

// IsViewer returns true if the role is viewer
func (wm *WorkspaceMember) IsViewer() bool {
	return wm.Role == string(RoleViewer)
}
