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

    // Init repository and handler
    barangRepo := repositories.NewBarangRepo(db)
    barangHandler := handlers.NewBarangHandler(barangRepo)

    // Router setup
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) })

    // API routes
    r.Route("/api", func(api chi.Router) {
        api.Get("/barang", barangHandler.GetAll)
        api.Get("/barang/{id}", barangHandler.GetByID)
        api.Post("/barang", barangHandler.Create)
        api.Put("/barang/{id}", barangHandler.UpdateBarang)
        api.Delete("/barang/{id}", barangHandler.DeleteBarang)
    })

    log.Println("Server listening on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
