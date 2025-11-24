package middleware

import (
    "context"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("super-secret-key")

type ctxKey string

const (
    ctxUserID ctxKey = "user_id"
    ctxRole   ctxKey = "role"
)

// AuthMiddleware verifies Bearer JWT and sets user_id and role into context.
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
            http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
            return
        }
        tokenStr := strings.TrimPrefix(auth, "Bearer ")
        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrTokenSignatureInvalid
            }
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "invalid token claims", http.StatusUnauthorized)
            return
        }
        // Check exp if present (jwt lib can also validate, but keep it simple)
        if exp, ok := claims["exp"].(float64); ok {
            if time.Now().Unix() > int64(exp) {
                http.Error(w, "token expired", http.StatusUnauthorized)
                return
            }
        }
        var userID int64
        if v, ok := claims["user_id"].(float64); ok {
            userID = int64(v)
        } else {
            http.Error(w, "missing user_id claim", http.StatusUnauthorized)
            return
        }
        role, _ := claims["role"].(string)
        ctx := context.WithValue(r.Context(), ctxUserID, userID)
        ctx = context.WithValue(ctx, ctxRole, role)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// UserIDFromContext retrieves user_id from request context.
func UserIDFromContext(ctx context.Context) (int64, bool) {
    v := ctx.Value(ctxUserID)
    id, ok := v.(int64)
    return id, ok
}

// RoleFromContext retrieves role from request context.
func RoleFromContext(ctx context.Context) (string, bool) {
    v := ctx.Value(ctxRole)
    s, ok := v.(string)
    return s, ok
}

// RequireRoles allows only specified roles to access a route.
func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
    allowedSet := make(map[string]struct{}, len(allowed))
    for _, a := range allowed { allowedSet[a] = struct{}{} }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role, _ := RoleFromContext(r.Context())
            if _, ok := allowedSet[role]; !ok {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
