package utils

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired   = errors.New("Token Expired")
	ErrTokenInvalid   = errors.New("Token Tidak Valid")
	ErrTokenMalformed = errors.New("Format Token Salah")
	ErrUnexpectedSign = errors.New("Signing Method Tidak Sesuai")
	ErrSecretNotSet   = errors.New("JWT Secret Belum di-set di Environment")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID   string    `json:"user_id"`
	UserName string    `json:"user_name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret []byte
	issuer       string
}

func NewJwtManager() (*JWTManager, error) {
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		return nil, ErrSecretNotSet // jangan pakai fallback default
	}
	return &JWTManager{
		accessSecret: []byte(secret),
		issuer:       "manage-app-auth-service", // issuer harus milik auth-service
	}, nil
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrUnexpectedSign
			}
			return m.accessSecret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(m.issuer),
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrTokenMalformed
		default:
			return nil, ErrTokenInvalid
		}
	}

	if !token.Valid || claims.Type != AccessToken {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
