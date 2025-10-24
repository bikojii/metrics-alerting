package handler

import (
	"github.com/admin/metrics-alerting/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

func UpdateHandler(store *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
		if len(parts) < 2 || parts[1] == "" {
			http.Error(w, "Not Fund", http.StatusNotFound)
			return
		}

		mType, name, valueStr := parts[0], parts[1], parts[2]

		switch mType {
		case "gauge":
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, "Bad Request: invalid gauge value", http.StatusBadRequest)
				return
			}
			store.UpdateGauge(name, value)

		case "counter":
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, "Bad Request: invalid counter value", http.StatusBadRequest)
				return
			}
			store.UpdateCounter(name, value)

		default:
			http.Error(w, "Bad Request: unknown metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
