package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	MemberID string `json:"mid"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTManager struct {
	secret          []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewJWTManager(secret string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secret:          []byte(secret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (m *JWTManager) GenerateAccessToken(memberID, role, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "gatie",
		},
		MemberID: memberID,
		Role:     role,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) RefreshDuration() time.Duration {
	return m.refreshDuration
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
