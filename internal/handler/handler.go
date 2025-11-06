package handler

import (
	"fmt"
	models "github.com/admin/metrics-alerting/internal/model"
	"github.com/admin/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
	"html"
	"net/http"
	"strconv"
)

func UpdateHandler(store *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")

		if mType == "" || name == "" || valueStr == "" {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}

		var metric models.Metrics

		switch mType {
		case models.Gauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, "Bad Request: invalid gauge value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    name,
				MType: models.Gauge,
				Value: &value,
			}

		case models.Counter:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, "Bad Request: invalid counter value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    name,
				MType: models.Counter,
				Delta: &value,
			}

		default:
			http.Error(w, "Bad Request: unknown metric type", http.StatusBadRequest)
			return
		}

		store.SaveMetric(metric)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func GetValueHandler(store *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		if mType == "" || name == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metric, found := store.GetMetric(name, mType)
		if !found {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		switch metric.MType {
		case models.Gauge:
			fmt.Fprintf(w, "%f", *metric.Value)

		case models.Counter:
			fmt.Fprintf(w, "%d", *metric.Delta)
		}
	}
}

func ListMetricsHandler(store *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "<html><body><h1>Metrics</h1><ul>")

		for name, value := range store.Gauges {
			fmt.Fprintf(w, "<li>%s = %f</li>", html.EscapeString(name), value)
		}

		for name, delta := range store.Counters {
			fmt.Fprintf(w, "<li>%s = %d</li>", html.EscapeString(name), delta)
		}

		fmt.Fprintf(w, "</ul></body></html>")
	}
}
