package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"database/sql"

	"warehouse/apperr"
)

type Meta struct {
    Page  int `json:"page,omitempty"`
    Limit int `json:"limit,omitempty"`
    Total int `json:"total,omitempty"`
}

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, resp APIResponse) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(resp)
}

func StatusFromError(err error) int {
    switch {
    case errors.Is(err, apperr.ErrInsufficientStock):
        return http.StatusBadRequest
    case errors.Is(err, apperr.ErrNotFound), errors.Is(err, sql.ErrNoRows):
        return http.StatusNotFound
    case errors.Is(err, apperr.ErrValidation):
        return http.StatusUnprocessableEntity
    default:
        return http.StatusInternalServerError
    }
}
