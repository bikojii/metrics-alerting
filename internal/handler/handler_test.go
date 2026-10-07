package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bikojii/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestHTTPAPI(t *testing.T) {
	store := repository.NewMemStorage()
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", UpdateHandler(store))
	r.Get("/value/{type}/{name}", GetValueHandler(store))
	r.Get("/", ListMetricsHandler(store))
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"POST", "/update/gauge/precise/3.1415926535", 200, "OK"},
		{"GET", "/value/gauge/precise", 200, "3.1415926535"},
		{"POST", "/update/gauge/a%2Fb/2.5", 200, "OK"},
		{"GET", "/value/gauge/a%2Fb", 200, "2.5"},
		{"POST", "/update/counter/polls/5", 200, "OK"},
		{"POST", "/update/counter/polls/2", 200, "OK"},
		{"GET", "/value/counter/polls", 200, "7"},
		{"GET", "/value/gauge/missing", 404, ""},
		{"GET", "/value/unknown/name", 400, ""},
		{"POST", "/update/gauge/name/NaN", 400, ""},
		{"POST", "/update/gauge/name/+Inf", 400, ""},
		{"POST", "/update/gauge/name/wrong", 400, ""},
		{"POST", "/update/counter/name/1.5", 400, ""},
		{"POST", "/update/unknown/name/1", 400, ""},
		{"POST", "/update/gauge//1", 404, ""},
		{"GET", "/update/gauge/name/1", 405, ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if tc.body != "" && w.Body.String() != tc.body {
				t.Fatalf("body = %q, want %q", w.Body.String(), tc.body)
			}
		})
	}
	store.UpdateGauge("<script>alert(1)</script>", 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatal("incorrect content type")
	}
	if strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "&lt;script&gt;") {
		t.Fatal("metric name is not escaped")
	}
}
