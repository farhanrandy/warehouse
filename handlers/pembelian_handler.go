package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    "strconv"

    "warehouse/models"
    "warehouse/repositories"
)

// PembelianHandler provides HTTP handler for purchase transactions (pembelian).
type PembelianHandler struct {
    Repo *repositories.PembelianRepo
}

func NewPembelianHandler(repo *repositories.PembelianRepo) *PembelianHandler {
    return &PembelianHandler{Repo: repo}
}

// CreatePembelianHandler handles POST /api/pembelian
// Expected JSON body maps to models.BeliHeader including an array of Details.
// Example:
// {
//   "no_faktur": "F-001",
//   "supplier": "PT Supplier",
//   "user_id": 1,
//   "details": [
//     {"barang_id": 10, "qty": 5, "harga": 12000},
//     {"barang_id": 12, "qty": 3, "harga": 15000}
//   ]
// }
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
