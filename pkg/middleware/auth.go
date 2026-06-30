package middleware

import (
	"strings"
	"time"

	"github.com/afifudin23/saepedia-api/config"
	"github.com/afifudin23/saepedia-api/pkg/jwt"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// Context keys.
const (
	CtxUserID     = "userID"
	CtxActiveRole = "activeRole"
	CtxJTI        = "tokenJTI"
	CtxTokenExp   = "tokenExp"
)

// revocationChecker dipasang sekali di router (logout denylist). Bila nil,
// pengecekan dilewati (mis. saat unit test).
var revocationChecker func(jti string) bool

// SetRevocationChecker memasang fungsi pengecek token yang sudah di-logout.
func SetRevocationChecker(fn func(jti string) bool) { revocationChecker = fn }

// Role constants — dipakai konsisten di seluruh aplikasi.
const (
	RoleAdmin  = "admin"
	RoleSeller = "seller"
	RoleBuyer  = "buyer"
	RoleDriver = "driver"
)

// Auth memverifikasi JWT dan menyimpan userID + activeRole ke context.
// Tidak peduli role apa — hanya memastikan token valid (user sudah login).
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header tidak ada")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" || tokenString == authHeader {
			response.Unauthorized(c, "format token harus 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := jwt.Verify(tokenString, config.AppConfig.AccessKey)
		if err != nil {
			response.Error(c, 401, "token tidak valid atau kadaluarsa", nil, response.ErrCodeTokenInvalid)
			c.Abort()
			return
		}

		// Tolak token yang sudah di-logout (ada di denylist).
		if revocationChecker != nil && claims.ID != "" && revocationChecker(claims.ID) {
			response.Error(c, 401, "token sudah tidak berlaku (logged out)", nil, response.ErrCodeTokenInvalid)
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UID)
		c.Set(CtxActiveRole, claims.ActiveRole)
		c.Set(CtxJTI, claims.ID)
		if claims.ExpiresAt != nil {
			c.Set(CtxTokenExp, claims.ExpiresAt.Time)
		}
		c.Next()
	}
}

// RequireRole memastikan ROLE AKTIF user termasuk salah satu role yang diizinkan.
// Authorization mengikuti role aktif, bukan seluruh role yang dimiliki user.
// Harus dipasang SETELAH Auth().
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		active := ActiveRole(c)
		if active == "" {
			response.Forbidden(c, "belum memilih role aktif, silakan pilih role dulu via /auth/select-role")
			c.Abort()
			return
		}

		for _, r := range roles {
			if active == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "role aktif '"+active+"' tidak diizinkan mengakses resource ini")
		c.Abort()
	}
}

// UserID mengambil id user dari context.
func UserID(c *gin.Context) string {
	if v, ok := c.Get(CtxUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ActiveRole mengambil role aktif dari context.
func ActiveRole(c *gin.Context) string {
	if v, ok := c.Get(CtxActiveRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// JTI mengambil id token (untuk logout).
func JTI(c *gin.Context) string {
	if v, ok := c.Get(CtxJTI); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// TokenExp mengambil waktu kadaluarsa token.
func TokenExp(c *gin.Context) time.Time {
	if v, ok := c.Get(CtxTokenExp); ok {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}
