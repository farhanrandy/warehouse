package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"warehouse/repositories"

	"github.com/go-chi/chi/v5"
)

// StokHandler provides endpoints for stock and history read operations.
type StokHandler struct {
    Repo *repositories.StokRepo
}

func NewStokHandler(repo *repositories.StokRepo) *StokHandler { return &StokHandler{Repo: repo} }

// standardResponse reused (defined in barang_handler.go); duplicate small struct for isolation.
type stokResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

// GET /api/stok
func (h *StokHandler) GetStokAkhirAll(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.Repo.GetStokAkhirAll(ctx)
    if err != nil {
        writeStokJSON(w, http.StatusInternalServerError, stokResponse{false, err.Error(), nil})
        return
    }
    writeStokJSON(w, http.StatusOK, stokResponse{true, "OK", list})
}

// GET /api/history-stok
func (h *StokHandler) GetHistoryAll(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.Repo.GetHistoryAll(ctx)
    if err != nil {
        writeStokJSON(w, http.StatusInternalServerError, stokResponse{false, err.Error(), nil})
        return
    }
    writeStokJSON(w, http.StatusOK, stokResponse{true, "OK", list})
}

// GET /api/stok/{barang_id}
func (h *StokHandler) GetStokByBarangHandler(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "barang_id")
    barangID, _ := strconv.ParseInt(idStr, 10, 64)
    if barangID <= 0 {
        writeStokJSON(w, http.StatusBadRequest, stokResponse{false, "invalid barang_id", nil})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    item, err := h.Repo.GetStokByBarangID(ctx, barangID)
    if err != nil {
        writeStokJSON(w, http.StatusInternalServerError, stokResponse{false, err.Error(), nil})
        return
    }
    if item == nil {
        writeStokJSON(w, http.StatusNotFound, stokResponse{false, "not found", nil})
        return
    }
    writeStokJSON(w, http.StatusOK, stokResponse{true, "OK", item})
}

// GET /api/history-stok/{barang_id}
func (h *StokHandler) GetHistoryByBarangHandler(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "barang_id")
    barangID, _ := strconv.ParseInt(idStr, 10, 64)
    if barangID <= 0 {
        writeStokJSON(w, http.StatusBadRequest, stokResponse{false, "invalid barang_id", nil})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.Repo.GetHistoryByBarangID(ctx, barangID)
    if err != nil {
        writeStokJSON(w, http.StatusInternalServerError, stokResponse{false, err.Error(), nil})
        return
    }
    writeStokJSON(w, http.StatusOK, stokResponse{true, "OK", list})
}

func writeStokJSON(w http.ResponseWriter, status int, payload stokResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
