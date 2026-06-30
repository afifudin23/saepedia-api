package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/auth/domain"
	userdomain "github.com/afifudin23/saepedia-api/internal/user/domain"
)

type UserData struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	IsAdmin   bool     `json:"is_admin"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

type AuthResponse struct {
	User              UserData `json:"user"`
	Token             string   `json:"token"`
	ActiveRole        string   `json:"active_role"`
	NeedRoleSelection bool     `json:"need_role_selection"`
}

// ProfileResponse dipakai oleh GET /auth/me dan halaman dashboard summary.
type ProfileResponse struct {
	User       UserData `json:"user"`
	ActiveRole string   `json:"active_role"`
}

func toUserData(u *userdomain.User) UserData {
	roles := u.Roles
	if roles == nil {
		roles = []string{}
	}
	return UserData{
		ID:        u.ID,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		Roles:     roles,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func ToAuthResponse(r *domain.AuthResult) AuthResponse {
	return AuthResponse{
		User:              toUserData(r.User),
		Token:             r.Token,
		ActiveRole:        r.ActiveRole,
		NeedRoleSelection: r.NeedRoleSelection,
	}
}

func ToProfileResponse(u *userdomain.User, activeRole string) ProfileResponse {
	return ProfileResponse{
		User:       toUserData(u),
		ActiveRole: activeRole,
	}
}
