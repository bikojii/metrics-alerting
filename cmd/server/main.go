package main

import (
	"github.com/admin/metrics-alerting/internal/handler"
	"github.com/admin/metrics-alerting/internal/repository"
	"log"
	"net/http"
)

func main() {
	store := repository.NewMemStorage()

	http.HandleFunc("/update/", handler.UpdateHandler(store))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
