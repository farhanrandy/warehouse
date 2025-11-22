package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"warehouse/models"
	"warehouse/repositories"
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
