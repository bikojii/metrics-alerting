package handler

import (
	"fmt"
	"html"
	"math"
	"net/http"
	"net/url"
	"strconv"

	models "github.com/bikojii/metrics-alerting/internal/model"
	"github.com/bikojii/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

func UpdateHandler(store repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		if r.URL.RawPath != "" {
			var err error
			name, err = url.PathUnescape(name)
			if err != nil {
				http.Error(w, "Bad Request: invalid metric name", http.StatusBadRequest)
				return
			}
		}
		valueStr := chi.URLParam(r, "value")

		if mType == "" || name == "" || valueStr == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		var metric models.Metrics

		switch mType {
		case models.Gauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
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

func GetValueHandler(store repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		if r.URL.RawPath != "" {
			var err error
			name, err = url.PathUnescape(name)
			if err != nil {
				http.Error(w, "Bad Request: invalid metric name", http.StatusBadRequest)
				return
			}
		}

		if mType == "" || name == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if mType != models.Gauge && mType != models.Counter {
			http.Error(w, "Bad Request: unknown metric type", http.StatusBadRequest)
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
			fmt.Fprint(w, strconv.FormatFloat(*metric.Value, 'f', -1, 64))

		case models.Counter:
			fmt.Fprintf(w, "%d", *metric.Delta)
		}
	}
}

func ListMetricsHandler(store repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>Metrics</title></head><body><h1>Metrics</h1><ul>")
		for _, metric := range store.ListMetrics() {
			var value string
			if metric.MType == models.Gauge {
				value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			} else {
				value = strconv.FormatInt(*metric.Delta, 10)
			}
			fmt.Fprintf(w, "<li>%s (%s) = %s</li>", html.EscapeString(metric.ID), metric.MType, value)
		}
		fmt.Fprint(w, "</ul></body></html>")
	}
}
