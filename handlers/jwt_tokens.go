package handlers

// Simple JWT helper functions for access & refresh tokens.
// Keep code minimal and beginner friendly.

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Reuse existing secret defined also in auth_handler.go and middleware.
var (
    accessTTL  = 15 * time.Minute
    refreshTTL = 7 * 24 * time.Hour
)

// GenerateAccessToken returns a short-lived JWT string.
func GenerateAccessToken(userID int64, role string) (string, error) {
    return generateToken(userID, role, accessTTL)
}

// GenerateRefreshToken returns a longer-lived JWT string.
func GenerateRefreshToken(userID int64, role string) (string, error) {
    return generateToken(userID, role, refreshTTL)
}

// Internal shared generator.
func generateToken(userID int64, role string, ttl time.Duration) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "role":    role,
        "exp":     time.Now().Add(ttl).Unix(),
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return t.SignedString(jwtSecret)
}

// ParseAndValidateAccess parses an access token and returns its claims.
func ParseAndValidateAccess(tokenStr string) (jwt.MapClaims, error) {
    return parseToken(tokenStr)
}

// ParseAndValidateRefresh parses a refresh token and returns its claims.
func ParseAndValidateRefresh(tokenStr string) (jwt.MapClaims, error) {
    return parseToken(tokenStr)
}

// Shared parsing logic.
func parseToken(tokenStr string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, jwt.ErrTokenMalformed
    }
    return claims, nil
}
