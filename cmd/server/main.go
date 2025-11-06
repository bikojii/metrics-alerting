package main

import (
	"github.com/bikojii/metrics-alerting/internal/config"
	"github.com/bikojii/metrics-alerting/internal/handler"
	"github.com/bikojii/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadServerConfig()

	store := repository.NewMemStorage()
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler(store))
	r.Get("/value/{type}/{name}", handler.GetValueHandler(store))
	r.Get("/", handler.ListMetricsHandler(store))

	log.Println("Server running on : ", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
