package dto

type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	// Roles: pilih satu atau lebih dari buyer/seller/driver. Default buyer bila kosong.
	Roles []string `json:"roles" binding:"omitempty,dive,oneof=buyer seller driver"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SelectRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin buyer seller driver"`
}
