package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bikojii/metrics-alerting/internal/config"
	"github.com/bikojii/metrics-alerting/internal/handler"
	"github.com/bikojii/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.LoadServerConfig()
	store := repository.NewMemStorage()
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler(store))
	r.Get("/value/{type}/{name}", handler.GetValueHandler(store))
	r.Get("/", handler.ListMetricsHandler(store))
	server := &http.Server{
		Addr: cfg.Address, Handler: r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Shutdown failed: %v", err)
			_ = server.Close()
		}
	}()
	log.Println("Server running on", cfg.Address)
	err := server.ListenAndServe()
	stop()
	<-done
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
