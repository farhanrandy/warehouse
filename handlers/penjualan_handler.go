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

// PenjualanHandler provides HTTP handler for sales transactions (penjualan).
type PenjualanHandler struct {
    Repo *repositories.PenjualanRepo
}

func NewPenjualanHandler(repo *repositories.PenjualanRepo) *PenjualanHandler {
    return &PenjualanHandler{Repo: repo}
}

// CreatePenjualanHandler handles POST /api/penjualan
// Body should map to models.JualHeader and include Details slice.
func (h *PenjualanHandler) CreatePenjualanHandler(w http.ResponseWriter, r *http.Request) {
    var hdr models.JualHeader
    if err := json.NewDecoder(r.Body).Decode(&hdr); err != nil {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid json", Data: nil})
        return
    }

    // Basic validation
    if hdr.NoFaktur == "" || hdr.Customer == "" || hdr.UserID <= 0 || len(hdr.Details) == 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "no_faktur, customer, user_id, details required", Data: nil})
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
    if err := h.Repo.CreatePenjualanTx(ctx, &hdr); err != nil {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusCreated, standardResponse{Success: true, Message: "created", Data: hdr})
}

// GetAll handles GET /api/penjualan
func (h *PenjualanHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.Repo.GetAll(ctx)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: list})
}

// penjualanDetailData structures header and details for explicit JSON fields.
type penjualanDetailData struct {
    Header  models.JualHeader   `json:"header"`
    Details []models.JualDetail `json:"details"`
}

// GetByID handles GET /api/penjualan/{id}
func (h *PenjualanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        writeJSON(w, http.StatusBadRequest, standardResponse{Success: false, Message: "invalid id", Data: nil})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    hdr, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, standardResponse{Success: false, Message: err.Error(), Data: nil})
        return
    }
    if hdr == nil {
        writeJSON(w, http.StatusNotFound, standardResponse{Success: false, Message: "not found", Data: nil})
        return
    }
    headerOnly := *hdr
    headerOnly.Details = nil
    payload := penjualanDetailData{Header: headerOnly, Details: hdr.Details}
    writeJSON(w, http.StatusOK, standardResponse{Success: true, Message: "OK", Data: payload})
}
