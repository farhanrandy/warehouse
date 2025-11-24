package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"warehouse/repositories"

	"github.com/go-chi/chi/v5"
)

type StokHandler struct {
    Repo *repositories.StokRepo
}

func NewStokHandler(repo *repositories.StokRepo) *StokHandler { return &StokHandler{Repo: repo} }

func (h *StokHandler) GetStokAkhirAll(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, err := h.Repo.GetStokAkhirAll(ctx)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: list})
}

func (h *StokHandler) GetHistoryAll(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    limit, _ := strconv.Atoi(q.Get("limit"))
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 10 }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, total, err := h.Repo.GetHistoryAll(ctx, page, limit)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: list, Meta: &Meta{Page: page, Limit: limit, Total: total}})
}

func (h *StokHandler) GetStokByBarangHandler(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "barang_id")
    barangID, _ := strconv.ParseInt(idStr, 10, 64)
    if barangID <= 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid barang_id"})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    item, err := h.Repo.GetStokByBarangID(ctx, barangID)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    if item == nil {
        WriteJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "not found"})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: item})
}

func (h *StokHandler) GetHistoryByBarangHandler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    limit, _ := strconv.Atoi(q.Get("limit"))
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 10 }

    idStr := chi.URLParam(r, "barang_id")
    barangID, _ := strconv.ParseInt(idStr, 10, 64)
    if barangID <= 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid barang_id"})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, total, err := h.Repo.GetHistoryByBarangID(ctx, barangID, page, limit)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: list, Meta: &Meta{Page: page, Limit: limit, Total: total}})
}
