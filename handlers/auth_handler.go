package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
    // Create token
    claims := jwt.MapClaims{
        "user_id": u.ID,
        "role":    u.Role,
        "exp":     time.Now().Add(1 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenStr, err := token.SignedString(jwtSecret)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "failed to sign token"})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Login success", Data: map[string]string{"token": tokenStr}})
}
