package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/bikojii/metrics-alerting/internal/config"
	models "github.com/bikojii/metrics-alerting/internal/model"
)

type Agent struct {
	mu             sync.Mutex
	reportMu       sync.Mutex
	client         *http.Client
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Metrics        map[string]*models.Metrics
}

func NewAgent(serverAddr string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		ServerAddr:     serverAddr,
		client:         &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }},
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Metrics: map[string]*models.Metrics{
			"PollCount":   {ID: "PollCount", MType: models.Counter, Delta: new(int64)},
			"RandomValue": {ID: "RandomValue", MType: models.Gauge, Value: new(float64)},
		},
	}
}

func (a *Agent) CollectMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	*a.Metrics["RandomValue"].Value = rand.Float64() * 100

	*a.Metrics["PollCount"].Delta += 1

	gaugeMap := map[string]float64{
		"Alloc":         float64(mem.Alloc),
		"HeapAlloc":     float64(mem.HeapAlloc),
		"HeapIdle":      float64(mem.HeapIdle),
		"HeapInuse":     float64(mem.HeapInuse),
		"HeapObjects":   float64(mem.HeapObjects),
		"HeapReleased":  float64(mem.HeapReleased),
		"HeapSys":       float64(mem.HeapSys),
		"StackInuse":    float64(mem.StackInuse),
		"StackSys":      float64(mem.StackSys),
		"Sys":           float64(mem.Sys),
		"Mallocs":       float64(mem.Mallocs),
		"Frees":         float64(mem.Frees),
		"NumGC":         float64(mem.NumGC),
		"LastGC":        float64(mem.LastGC),
		"PauseTotalNs":  float64(mem.PauseTotalNs),
		"GCCPUFraction": mem.GCCPUFraction,
	}

	for name, value := range gaugeMap {
		v := value
		a.Metrics[name] = &models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		}
	}
}

func (a *Agent) SendMetric(m *models.Metrics) error {
	return a.sendMetric(context.Background(), m)
}

func (a *Agent) sendMetric(ctx context.Context, m *models.Metrics) error {
	if m == nil || m.ID == "" {
		return fmt.Errorf("metric name is required")
	}
	var value string
	switch m.MType {
	case models.Gauge:
		if m.Value == nil || math.IsNaN(*m.Value) || math.IsInf(*m.Value, 0) {
			return fmt.Errorf("invalid gauge value")
		}
		value = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case models.Counter:
		if m.Delta == nil {
			return fmt.Errorf("counter delta is required")
		}
		value = strconv.FormatInt(*m.Delta, 10)
	default:
		return fmt.Errorf("unknown metric type: %s", m.MType)
	}
	address, err := config.NormalizeServerAddress(a.ServerAddr)
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%s/update/%s/%s/%s", address, m.MType, url.PathEscape(m.ID), value)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

func (a *Agent) Report(ctx context.Context) error {
	a.reportMu.Lock()
	defer a.reportMu.Unlock()
	a.mu.Lock()
	metrics := make([]models.Metrics, 0, len(a.Metrics))
	for _, metric := range a.Metrics {
		m := *metric
		if m.Value != nil {
			value := *m.Value
			m.Value = &value
		}
		if m.Delta != nil {
			delta := *m.Delta
			m.Delta = &delta
		}
		metrics = append(metrics, m)
	}
	a.mu.Unlock()
	var errs []error
	for _, m := range metrics {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}
		if m.MType == models.Counter && m.Delta != nil && *m.Delta == 0 {
			continue
		}
		if err := a.sendMetric(ctx, &m); err != nil {
			errs = append(errs, fmt.Errorf("send %s: %w", m.ID, err))
			continue
		}
		if m.MType == models.Counter {
			a.mu.Lock()
			*a.Metrics[m.ID].Delta -= *m.Delta
			a.mu.Unlock()
		}
	}
	return errors.Join(errs...)
}

func (a *Agent) Run(ctx context.Context) error {
	if a.PollInterval <= 0 || a.ReportInterval <= 0 {
		return fmt.Errorf("intervals must be positive")
	}
	if _, err := config.NormalizeServerAddress(a.ServerAddr); err != nil {
		return err
	}
	ticker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportInterval)
	defer ticker.Stop()
	defer reportTicker.Stop()
	var workers sync.WaitGroup
	workers.Add(1)
	defer workers.Wait()
	go func() {
		defer workers.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-reportTicker.C:
				if err := a.Report(ctx); err != nil && ctx.Err() == nil {
					log.Printf("Report failed: %v", err)
				}
			}
		}
	}()
	a.CollectMetrics()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.CollectMetrics()
		}
	}
}

func main() {
	cfg := config.LoadAgentConfig()

	log.Println("Agent will send metrics to", cfg.ServerAddress)
	log.Println("Report interval:", cfg.ReportInterval, "seconds")
	log.Println("Poll interval:", cfg.PollInterval, "seconds")

	a := NewAgent(
		cfg.ServerAddress,
		time.Duration(cfg.PollInterval)*time.Second,
		time.Duration(cfg.ReportInterval)*time.Second,
	)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := a.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
