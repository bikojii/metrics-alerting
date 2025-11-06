package main

import (
	"github.com/admin/metrics-alerting/internal/handler"
	"github.com/admin/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	store := repository.NewMemStorage()
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler(store))
	r.Get("/value/{type}/{name}", handler.GetValueHandler(store))
	r.Get("/", handler.ListMetricsHandler(store))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
