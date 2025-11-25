package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"warehouse/middleware"
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
func (h *PenjualanHandler) CreatePenjualanHandler(w http.ResponseWriter, r *http.Request) {
    var hdr models.JualHeader
    if err := json.NewDecoder(r.Body).Decode(&hdr); err != nil {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid json"})
        return
    }

    if hdr.Customer == "" || len(hdr.Details) == 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "customer, details required"})
        return
    }
    for i, d := range hdr.Details {
        if d.BarangID <= 0 || d.Qty <= 0 {
            WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid detail at index " + strconv.Itoa(i)})
            return
        }
    }

    // Set user from JWT context, ignore any user_id in body
    if uid, ok := middleware.UserIDFromContext(r.Context()); ok {
        hdr.UserID = uid
    } else {
        WriteJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Message: "unauthorized"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()
    if err := h.Repo.CreatePenjualanTx(ctx, &hdr); err != nil {
        code := StatusFromError(err)
        WriteJSON(w, code, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusCreated, APIResponse{Success: true, Message: "created", Data: hdr})
}

// GetAll handles GET /api/penjualan
func (h *PenjualanHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    limit, _ := strconv.Atoi(q.Get("limit"))
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 10 }

    var fromPtr, toPtr *time.Time
    if fs := q.Get("from"); fs != "" {
        if t, err := time.Parse("2006-01-02", fs); err == nil {
            fromPtr = &t
        }
    }
    if ts := q.Get("to"); ts != "" {
        if t, err := time.Parse("2006-01-02", ts); err == nil {
            t2 := t.Add(24*time.Hour - time.Nanosecond)
            toPtr = &t2
        }
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    list, total, err := h.Repo.GetAll(ctx, fromPtr, toPtr, page, limit)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: list, Meta: &Meta{Page: page, Limit: limit, Total: total}})
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
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid id"})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    hdr, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    if hdr == nil {
        WriteJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "not found"})
        return
    }
    headerOnly := *hdr
    headerOnly.Details = nil
    payload := penjualanDetailData{Header: headerOnly, Details: hdr.Details}
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: payload})
}
