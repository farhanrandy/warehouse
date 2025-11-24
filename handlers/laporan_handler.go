package handlers

import (
    "context"
    "net/http"
    "time"

    "warehouse/repositories"
)

type LaporanHandler struct {
    StokRepo       *repositories.StokRepo
    PenjualanRepo  *repositories.PenjualanRepo
    PembelianRepo  *repositories.PembelianRepo
}

func NewLaporanHandler(s *repositories.StokRepo, pj *repositories.PenjualanRepo, pb *repositories.PembelianRepo) *LaporanHandler {
    return &LaporanHandler{StokRepo: s, PenjualanRepo: pj, PembelianRepo: pb}
}

// GET /api/laporan/stok
func (h *LaporanHandler) LaporanStok(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.StokRepo.GetStokAkhirAll(ctx)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Laporan stok", Data: list})
}

// GET /api/laporan/penjualan?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *LaporanHandler) LaporanPenjualan(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    var fromPtr, toPtr *time.Time
    if fs := q.Get("from"); fs != "" {
        if t, err := time.Parse("2006-01-02", fs); err == nil { fromPtr = &t }
    }
    if ts := q.Get("to"); ts != "" {
        if t, err := time.Parse("2006-01-02", ts); err == nil {
            t2 := t.Add(24*time.Hour - time.Nanosecond)
            toPtr = &t2
        }
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.PenjualanRepo.GetReport(ctx, fromPtr, toPtr)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Laporan penjualan", Data: list})
}

// GET /api/laporan/pembelian?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *LaporanHandler) LaporanPembelian(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    var fromPtr, toPtr *time.Time
    if fs := q.Get("from"); fs != "" {
        if t, err := time.Parse("2006-01-02", fs); err == nil { fromPtr = &t }
    }
    if ts := q.Get("to"); ts != "" {
        if t, err := time.Parse("2006-01-02", ts); err == nil {
            t2 := t.Add(24*time.Hour - time.Nanosecond)
            toPtr = &t2
        }
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.PembelianRepo.GetReport(ctx, fromPtr, toPtr)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Laporan pembelian", Data: list})
}
