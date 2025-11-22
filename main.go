package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"warehouse/config"
	"warehouse/handlers"
	"warehouse/repositories"
)

func main() {
    // Open database connection from .env
    db, err := config.OpenDB()
    if err != nil {
        log.Fatalf("db connection error: %v", err)
    }
    defer db.Close()

    // Init repositories and handlers
    barangRepo := repositories.NewBarangRepo(db)
    barangHandler := handlers.NewBarangHandler(barangRepo)
    stokRepo := repositories.NewStokRepo(db)
    stokHandler := handlers.NewStokHandler(stokRepo)
    pembelianRepo := repositories.NewPembelianRepo(db)
    pembelianHandler := handlers.NewPembelianHandler(pembelianRepo)
    penjualanRepo := repositories.NewPenjualanRepo(db)
    penjualanHandler := handlers.NewPenjualanHandler(penjualanRepo)

    // Router setup
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) })

    // API routes
    r.Route("/api", func(api chi.Router) {
        // Master Barang CRUD
        api.Get("/barang", barangHandler.GetAll)
        api.Get("/barang/{id}", barangHandler.GetByID)
        api.Post("/barang", barangHandler.Create)
        api.Put("/barang/{id}", barangHandler.UpdateBarang)
        api.Delete("/barang/{id}", barangHandler.DeleteBarang)

        // Stok Management (read-only for now)
        api.Get("/stok", stokHandler.GetStokAkhirAll)
        api.Get("/history-stok", stokHandler.GetHistoryAll)
        api.Get("/stok/{barang_id}", stokHandler.GetStokByBarangHandler)
        api.Get("/history-stok/{barang_id}", stokHandler.GetHistoryByBarangHandler)

        // Transaksi Pembelian
        api.Post("/pembelian", pembelianHandler.CreatePembelianHandler)

        // Transaksi Penjualan
        api.Post("/penjualan", penjualanHandler.CreatePenjualanHandler)
    })

    log.Println("Server listening on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
