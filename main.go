package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"warehouse/config"
	"warehouse/handlers"
	wm "warehouse/middleware"
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
    userRepo := repositories.NewUserRepo(db)
    authHandler := handlers.NewAuthHandler(userRepo)
    laporanHandler := handlers.NewLaporanHandler(stokRepo, penjualanRepo, pembelianRepo)

    // Router setup
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) })

    // API routes
    r.Route("/api", func(api chi.Router) {
        // Public
        api.Post("/login", authHandler.Login)

        // Protected group
        api.Group(func(priv chi.Router) {
            priv.Use(wm.AuthMiddleware)

            // Master Barang CRUD
            priv.Get("/barang", barangHandler.GetAll)
            priv.Get("/barang/stok", barangHandler.GetAllWithStok)
            priv.Get("/barang/{id}", barangHandler.GetByID)
            priv.Post("/barang", barangHandler.Create)
            priv.Put("/barang/{id}", barangHandler.UpdateBarang)
            // Only admin can delete
            priv.With(wm.RequireRoles("admin")).Delete("/barang/{id}", barangHandler.DeleteBarang)

            // Stok and History
            priv.Get("/stok", stokHandler.GetStokAkhirAll)
            priv.Get("/history-stok", stokHandler.GetHistoryAll)
            priv.Get("/stok/{barang_id}", stokHandler.GetStokByBarangHandler)
            priv.Get("/history-stok/{barang_id}", stokHandler.GetHistoryByBarangHandler)

            // Transaksi Pembelian
            priv.Post("/pembelian", pembelianHandler.CreatePembelianHandler)
            priv.Get("/pembelian", pembelianHandler.GetAll)
            priv.Get("/pembelian/{id}", pembelianHandler.GetByID)

            // Transaksi Penjualan
            priv.Post("/penjualan", penjualanHandler.CreatePenjualanHandler)
            priv.Get("/penjualan", penjualanHandler.GetAll)
            priv.Get("/penjualan/{id}", penjualanHandler.GetByID)

            // Laporan
            priv.Get("/laporan/stok", laporanHandler.LaporanStok)
            priv.Get("/laporan/penjualan", laporanHandler.LaporanPenjualan)
            priv.Get("/laporan/pembelian", laporanHandler.LaporanPembelian)
        })
    })

    log.Println("Server listening on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
