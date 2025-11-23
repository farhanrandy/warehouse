package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"warehouse/models"
	"warehouse/repositories"

	"github.com/go-chi/chi/v5"
)

// standardResponse is the uniform JSON envelope for all responses.
type standardResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    interface{} `json:"meta,omitempty"`
}

// BarangHandler provides HTTP handlers for master_barang.
type BarangHandler struct {
    Repo *repositories.BarangRepo
}

func NewBarangHandler(repo *repositories.BarangRepo) *BarangHandler {
    return &BarangHandler{Repo: repo}
}

// GET /api/barang
func (h *BarangHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    // Beginner-friendly: parse query params with defaults
    q := r.URL.Query()
    search := q.Get("search")             // optional search keyword
    page, _ := strconv.Atoi(q.Get("page")) // default to 1 if missing/invalid
    limit, _ := strconv.Atoi(q.Get("limit")) // default to 10 if missing/invalid
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 10 }

    // Create a short-lived context for the DB call
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // Call the repository to get data and total rows for pagination
    items, total, err := h.Repo.GetAll(ctx, search, page, limit)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }

    // Compose meta information for pagination
    type meta struct {
        Page  int `json:"page"`
        Limit int `json:"limit"`
        Total int `json:"total"`
    }
    m := meta{Page: page, Limit: limit, Total: total}

    // Return standard response: data + meta side-by-side
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: items, Meta: m})
}

// GET /api/barang/{id}
func (h *BarangHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    // Read id from chi path params: /api/barang/{id}
    idStr := chi.URLParam(r, "id")
    if idStr == "" {
        // last fallback: accept id from query parameter if not present in path
        idStr = r.URL.Query().Get("id")
    }
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid id", Data: nil})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    item, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    if item == nil {
        writeJSON(w, http.StatusNotFound, standardResponse{Success: false, Message: "not found", Data: nil})
        return
    }
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: item})
}

// POST /api/barang
func (h *BarangHandler) Create(w http.ResponseWriter, r *http.Request) {
    var b models.Barang
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid json", Data: nil})
        return
    }
    // Minimal validation for beginners: ensure key fields are present
    if b.KodeBarang == "" || b.NamaBarang == "" || b.Satuan == "" {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "kode_barang, nama_barang, satuan are required", Data: nil})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Create(ctx, &b); err != nil {
        // Translate sql.ErrNoRows and other errors to messages
        if err == sql.ErrNoRows {
            writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "failed to create", Data: nil})
            return
        }
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusCreated, standardResponse{Success: true, Message: "created", Data: b})
}

// PUT /api/barang/{id}
func (h *BarangHandler) UpdateBarang(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid id", Data: nil})
        return
    }

    var b models.Barang
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid json", Data: nil})
        return
    }
    // Enforce ID from path to avoid mismatch
    b.ID = id
    // Basic validation
    if b.KodeBarang == "" || b.NamaBarang == "" || b.Satuan == "" {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "kode_barang, nama_barang, satuan are required", Data: nil})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Update(ctx, &b); err != nil {
        if err == sql.ErrNoRows {
            writeJSON(w, http.StatusNotFound, standardResponse{Success: false, Message: "not found", Data: nil})
            return
        }
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    // Optionally re-fetch to return latest state
    updated, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "updated", Data: updated})
}

// DELETE /api/barang/{id}
func (h *BarangHandler) DeleteBarang(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid id", Data: nil})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Delete(ctx, id); err != nil {
        if err == sql.ErrNoRows {
            writeJSON(w, http.StatusNotFound, standardResponse{Success: false, Message: "not found", Data: nil})
            return
        }
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "deleted", Data: map[string]int64{"id": id}})
}

func writeJSON(w http.ResponseWriter, status int, payload standardResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
