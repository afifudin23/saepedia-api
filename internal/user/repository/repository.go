package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/afifudin23/saepedia-api/internal/user/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User, roles []string) error {
	model := &UserModel{
		Email:    user.Email,
		Password: user.Password,
		IsAdmin:  user.IsAdmin,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			return mapUniqueErr(err)
		}
		for _, role := range roles {
			if err := tx.Create(&UserRoleModel{UserID: model.ID, Role: role}).Error; err != nil {
				return mapUniqueErr(err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	user.ID = model.ID
	user.Roles = roles
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *userRepository) findOne(ctx context.Context, query string, arg any) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Preload("Roles").Where(query, arg).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *userRepository) AddRole(ctx context.Context, userID, role string) error {
	err := r.db.WithContext(ctx).Create(&UserRoleModel{UserID: userID, Role: role}).Error
	if err != nil && isUniqueViolation(err) {
		return domain.ErrRoleAlreadyOwned
	}
	return err
}

func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&count).Error
	return count, err
}

// mapUniqueErr menerjemahkan error unique constraint ke error domain yang jelas.
func mapUniqueErr(err error) error {
	if !isUniqueViolation(err) {
		return err
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "email") {
		return domain.ErrEmailExists
	}
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
