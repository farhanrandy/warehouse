package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"warehouse/models"
	"warehouse/repositories"

	"github.com/go-chi/chi/v5"
)

// PembelianHandler provides HTTP handler for purchase transactions (pembelian).
type PembelianHandler struct {
    Repo *repositories.PembelianRepo
}

func NewPembelianHandler(repo *repositories.PembelianRepo) *PembelianHandler {
    return &PembelianHandler{Repo: repo}
}

// CreatePembelianHandler handles POST /api/pembelian
func (h *PembelianHandler) CreatePembelianHandler(w http.ResponseWriter, r *http.Request) {
    var hdr models.BeliHeader
    if err := json.NewDecoder(r.Body).Decode(&hdr); err != nil {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid json", Data: nil})
        return
    }

    // Basic validation before hitting DB.
    if hdr.NoFaktur == "" || hdr.Supplier == "" || hdr.UserID <= 0 || len(hdr.Details) == 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "no_faktur, supplier, user_id, details required", Data: nil})
        return
    }
    for i, d := range hdr.Details {
        if d.BarangID <= 0 || d.Qty <= 0 || d.Harga < 0 {
            writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid detail at index " + strconv.Itoa(i), Data: nil})
            return
        }
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()
    if err := h.Repo.CreatePembelianTx(ctx, &hdr); err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusCreated, standardResponse{Success: true, Message: "created", Data: hdr})
}

// GetAll handles GET /api/pembelian
func (h *PembelianHandler) GetAll(w http.ResponseWriter, r *http.Request) {
  
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    list, err := h.Repo.GetAll(ctx)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: list})
}

type pembelianDetailData struct {
    Header  models.BeliHeader   `json:"header"`
    Details []models.BeliDetail `json:"details"`
}

// GetByID handles GET /api/pembelian/{id}
func (h *PembelianHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid id", Data: nil})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // repository fetches header and details
    hdr, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    if hdr == nil {
        writeJSON(w, http.StatusNotFound, standardResponse{Success: false, Message: "not found", Data: nil})
        return
    }

    // prepare payload with separated header and details
    headerOnly := *hdr
    headerOnly.Details = nil
    payload := pembelianDetailData{Header: headerOnly, Details: hdr.Details}
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: payload})
}
