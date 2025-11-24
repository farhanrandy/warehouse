package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "golang.org/x/crypto/bcrypt"

    "warehouse/repositories"
)

type AuthHandler struct {
    Users *repositories.UserRepo
}

func NewAuthHandler(users *repositories.UserRepo) *AuthHandler { return &AuthHandler{Users: users} }

type loginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type refreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

// jwtSecret shared with middleware; keep simple by duplicating constant if needed.
var jwtSecret = []byte("super-secret-key")

// POST /api/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "invalid json"})
        return
    }
    if req.Username == "" || req.Password == "" {
        WriteJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "username and password required"})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    u, err := h.Users.GetByUsername(ctx, req.Username)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    if u == nil {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "invalid credentials"})
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "invalid credentials"})
        return
    }
    // Generate access (15m) & refresh (7d) tokens.
    accessToken, err := GenerateAccessToken(u.ID, u.Role)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "failed to generate access token"})
        return
    }
    refreshToken, err := GenerateRefreshToken(u.ID, u.Role)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "failed to generate refresh token"})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Login success", Data: map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    }})
}

// POST /api/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    var req refreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "invalid json"})
        return
    }
    if req.RefreshToken == "" {
        WriteJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "refresh_token required"})
        return
    }
    claims, err := ParseAndValidateRefresh(req.RefreshToken)
    if err != nil {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "invalid refresh token"})
        return
    }
    expVal, ok := claims["exp"].(float64)
    if !ok || time.Unix(int64(expVal), 0).Before(time.Now()) {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "refresh token expired"})
        return
    }
    uidFloat, okUID := claims["user_id"].(float64)
    role, okRole := claims["role"].(string)
    if !okUID || !okRole {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "invalid token claims"})
        return
    }
    accessToken, err := GenerateAccessToken(int64(uidFloat), role)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "failed to generate access token"})
        return
    }
    newRefresh, err := GenerateRefreshToken(int64(uidFloat), role)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "failed to generate refresh token"})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Token refreshed", Data: map[string]string{
        "access_token":  accessToken,
        "refresh_token": newRefresh,
    }})
}
