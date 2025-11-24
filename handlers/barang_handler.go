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

// BarangHandler provides HTTP handlers for master_barang.
type BarangHandler struct {
    Repo *repositories.BarangRepo
}

func NewBarangHandler(repo *repositories.BarangRepo) *BarangHandler {
    return &BarangHandler{Repo: repo}
}

// GET /api/barang
func (h *BarangHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    search := q.Get("search")             
    page, _ := strconv.Atoi(q.Get("page")) 
    limit, _ := strconv.Atoi(q.Get("limit")) 
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 10 }

    // Create a short-lived context for the DB call
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // Call the repository to get data and total rows for pagination
    items, total, err := h.Repo.GetAll(ctx, search, page, limit)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
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
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: items, Meta: &Meta{Page: m.Page, Limit: m.Limit, Total: m.Total}})
}

// GET /api/barang/{id}
func (h *BarangHandler) GetByID(w http.ResponseWriter, r *http.Request) {

    idStr := chi.URLParam(r, "id")
    if idStr == "" {
    
        idStr = r.URL.Query().Get("id")
    }
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid id"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    item, err := h.Repo.GetWithStokByID(ctx, id)
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

// POST /api/barang
func (h *BarangHandler) Create(w http.ResponseWriter, r *http.Request) {
    var b models.Barang
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid json"})
        return
    }

    if b.KodeBarang == "" || b.NamaBarang == "" || b.Satuan == "" {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "kode_barang, nama_barang, satuan are required"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Create(ctx, &b); err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusCreated, APIResponse{Success: true, Message: "created", Data: b})
}

// PUT /api/barang/{id}
func (h *BarangHandler) UpdateBarang(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid id"})
        return
    }

    var b models.Barang
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid json"})
        return
    }
    // Enforce ID from path to avoid mismatch
    b.ID = id

    if b.KodeBarang == "" || b.NamaBarang == "" || b.Satuan == "" {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "kode_barang, nama_barang, satuan are required"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Update(ctx, &b); err != nil {
        if err == sql.ErrNoRows {
            WriteJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "not found"})
            return
        }
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
 
    updated, err := h.Repo.GetByID(ctx, id)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "updated", Data: updated})
}

// DELETE /api/barang/{id}
func (h *BarangHandler) DeleteBarang(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if id <= 0 {
        WriteJSON(w, http.StatusUnprocessableEntity, APIResponse{Success: false, Message: "invalid id"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := h.Repo.Delete(ctx, id); err != nil {
        if err == sql.ErrNoRows {
            WriteJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "not found"})
            return
        }
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "deleted", Data: map[string]int64{"id": id}})
}

// uses WriteJSON from response.go

// GET /api/barang/stok
func (h *BarangHandler) GetAllWithStok(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    items, err := h.Repo.GetAllWithStok(ctx)
    if err != nil {
        WriteJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
        return
    }
    WriteJSON(w, http.StatusOK, APIResponse{Success: true, Message: "OK", Data: items})
}
