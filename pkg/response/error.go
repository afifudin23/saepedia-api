package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Daftar ErrorCode yang tersedia.
const (
	ErrCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeUnprocessable       ErrorCode = "UNPROCESSABLE_ENTITY"
	ErrCodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrCodeTokenExpired        ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid        ErrorCode = "TOKEN_INVALID"
)

// Error mengembalikan response error dengan status code custom.
func Error(c *gin.Context, statusCode int, message string, err any, errorCode ErrorCode) {
	c.JSON(statusCode, Response{
		Status:    false,
		Data:      nil,
		Message:   message,
		Error:     err,
		ErrorCode: errorCode,
	})
}

// ValidationError mengembalikan 400 dengan detail field yang gagal validasi.
func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		BadRequest(c, err.Error())
		return
	}

	fields := make(map[string]string)
	for _, fe := range ve {
		fields[toSnakeCase(fe.Field())] = validationMessage(fe)
	}

	c.JSON(http.StatusBadRequest, Response{
		Status:    false,
		Data:      nil,
		Message:   "Validation failed",
		Error:     fields,
		ErrorCode: ErrCodeValidation,
	})
}

// BadRequest mengembalikan 400.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, nil, ErrCodeBadRequest)
}

// Unauthorized mengembalikan 401.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil, ErrCodeUnauthorized)
}

// Forbidden mengembalikan 403.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil, ErrCodeForbidden)
}

// NotFound mengembalikan 404.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil, ErrCodeNotFound)
}

// Conflict mengembalikan 409.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message, nil, ErrCodeConflict)
}

// UnprocessableEntity mengembalikan 422 (mis. saldo kurang, stok habis).
func UnprocessableEntity(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message, nil, ErrCodeUnprocessable)
}

// InternalServerError mengembalikan 500 tanpa membocorkan detail error.
func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "Internal server error", nil, ErrCodeInternalServerError)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func validationMessage(fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, fe.Param())
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// toSnakeCase mengkonversi "FieldName" → "field_name".
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
