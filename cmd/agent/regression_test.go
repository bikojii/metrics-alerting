package main

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bikojii/metrics-alerting/internal/handler"
	models "github.com/bikojii/metrics-alerting/internal/model"
	"github.com/bikojii/metrics-alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestAgentServerReports(t *testing.T) {
	store := repository.NewMemStorage()
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler(store))
	testHandler := (r)
	a := NewAgent(strings.TrimPrefix("http://localhost:8080", "http://"), time.Second, time.Second)
	a.client.Transport = handlerTransport{handler: testHandler}
	for i := 0; i < 5; i++ {
		a.CollectMetrics()
	}
	if err := a.Report(context.Background()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		a.CollectMetrics()
	}
	if err := a.Report(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := a.Report(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.GetCounter("PollCount"); got != 10 {
		t.Fatalf("PollCount = %d, want 10", got)
	}
	if len(store.ListMetrics()) != 18 {
		t.Fatalf("metrics = %d, want 18", len(store.ListMetrics()))
	}
}

func TestReportPreservesFailedDelta(t *testing.T) {
	testHandler := (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	a := NewAgent("http://localhost:8080", time.Second, time.Second)
	a.client.Transport = handlerTransport{handler: testHandler}
	a.CollectMetrics()
	if err := a.Report(context.Background()); err == nil {
		t.Fatal("expected HTTP error")
	}
	if got := *a.Metrics["PollCount"].Delta; got != 1 {
		t.Fatalf("pending delta = %d", got)
	}
}

func TestSendMetricRequest(t *testing.T) {
	testHandler := (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/update/counter/Poll Count/7" || r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	a := NewAgent("http://localhost:8080", time.Second, time.Second)
	a.client.Transport = handlerTransport{handler: testHandler}
	d := int64(7)
	if err := a.SendMetric(&models.Metrics{ID: "Poll Count", MType: models.Counter, Delta: &d}); err != nil {
		t.Fatal(err)
	}
	for _, m := range []*models.Metrics{nil, {ID: "x", MType: models.Gauge}, {ID: "x", MType: "unknown"}} {
		if a.SendMetric(m) == nil {
			t.Fatal("accepted invalid metric")
		}
	}
}

func TestRunCollectsDuringSlowReport(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	testHandler := (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(started) })
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer close(release)
	a := NewAgent("http://localhost:8080", 5*time.Millisecond, 5*time.Millisecond)
	a.client.Transport = handlerTransport{handler: testHandler}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("report did not start")
	}
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		a.mu.Lock()
		count := *a.Metrics["PollCount"].Delta
		a.mu.Unlock()
		if count >= 4 {
			break
		}
		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatal("collection blocked by report")
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("agent did not stop")
	}
}

func TestSendMetricTimeout(t *testing.T) {
	testHandler := (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	a := NewAgent("http://localhost:8080", time.Second, time.Second)
	a.client.Transport = handlerTransport{handler: testHandler}
	a.client.Timeout = 20 * time.Millisecond
	a.CollectMetrics()
	if err := a.SendMetric(a.Metrics["RandomValue"]); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestReportKeepsCollectionsDuringSend(t *testing.T) {
	a := NewAgent("localhost:8080", time.Second, time.Second)
	a.CollectMetrics()
	a.client.Transport = handlerTransport{handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/counter/") {
			a.CollectMetrics()
		}
	})}
	if err := a.Report(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := *a.Metrics["PollCount"].Delta; got != 1 {
		t.Fatalf("pending delta = %d, want 1", got)
	}
}

func TestSendMetricRejectsRedirect(t *testing.T) {
	a := NewAgent("localhost:8080", time.Second, time.Second)
	a.client.Transport = handlerTransport{handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/elsewhere")
		w.WriteHeader(http.StatusFound)
	})}
	a.CollectMetrics()
	if err := a.SendMetric(a.Metrics["RandomValue"]); err == nil {
		t.Fatal("redirect counted as success")
	}
}

func TestRunRejectsInvalidConfig(t *testing.T) {
	for _, a := range []*Agent{
		NewAgent("localhost:8080", 0, time.Second),
		NewAgent("ftp://localhost", time.Second, time.Second),
	} {
		if err := a.Run(context.Background()); err == nil {
			t.Fatal("accepted invalid configuration")
		}
	}
}
