package repository

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/user/domain"
)

type UserModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	IsAdmin   bool   `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Roles []UserRoleModel `gorm:"foreignKey:UserID"`
}

func (UserModel) TableName() string { return "users" }

type UserRoleModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string `gorm:"type:uuid;not null;index"`
	Role      string `gorm:"not null"`
	CreatedAt time.Time
}

func (UserRoleModel) TableName() string { return "user_roles" }

func (m UserModel) toDomain() *domain.User {
	roles := make([]string, 0, len(m.Roles))
	for _, r := range m.Roles {
		roles = append(roles, r.Role)
	}
	return &domain.User{
		ID:        m.ID,
		Email:     m.Email,
		Password:  m.Password,
		IsAdmin:   m.IsAdmin,
		Roles:     roles,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
