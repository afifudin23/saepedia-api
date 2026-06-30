package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTTL adalah masa berlaku token. Didokumentasikan di README (session behavior).
const TokenTTL = 24 * time.Hour

// Claims menyimpan id user dan ROLE AKTIF yang dipilih untuk sesi ini.
// Authorization SELALU mengacu ke ActiveRole, bukan seluruh role yang dimiliki user.
type Claims struct {
	UID        string `json:"uid"`
	ActiveRole string `json:"active_role"`
	jwt.RegisteredClaims
}

// Generate membuat token dengan user id dan role aktif.
// activeRole boleh kosong jika user belum memilih role (punya >1 role non-admin).
func Generate(userID, activeRole, key string) (string, error) {
	jti, err := newJTI()
	if err != nil {
		return "", err
	}
	claims := Claims{
		UID:        userID,
		ActiveRole: activeRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // dipakai untuk denylist saat logout
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

// newJTI menghasilkan id token acak (16 byte hex).
func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Verify(tokenString, key string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method tidak valid")
			}
			return []byte(key), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}

	return claims, nil
}
