package main

import (
	"fmt"
	"github.com/bikojii/metrics-alerting/internal/config"
	models "github.com/bikojii/metrics-alerting/internal/model"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type Agent struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Metrics        map[string]*models.Metrics
}

func NewAgent(serverAddr string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		ServerAddr:     serverAddr,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Metrics: map[string]*models.Metrics{
			"PollCount":   {ID: "PollCount", MType: models.Counter, Delta: new(int64)},
			"RandomValue": {ID: "RandomValue", MType: models.Gauge, Value: new(float64)},
		},
	}
}

func (a *Agent) CollectMetrics() {
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
	var valueStr string
	if m.MType == models.Gauge && m.Value != nil {
		valueStr = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	} else if m.MType == models.Counter && m.Delta != nil {
		valueStr = strconv.FormatInt(*m.Delta, 10)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", a.ServerAddr, m.MType, m.ID, valueStr)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ПРОЧИТАТЬ ПОДРОБНЕЕ
func (a *Agent) Run() {
	ticker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportInterval)
	defer ticker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-ticker.C:
			a.CollectMetrics()
		case <-reportTicker.C:
			for _, m := range a.Metrics {
				_ = a.SendMetric(m)
			}
		}
	}
}

func main() {
	cfg := config.LoadAgentConfig()

	log.Println("Agent will send metrics to", cfg.ServerAddress)
	log.Println("Report interval:", cfg.ReportInterval, "seconds")
	log.Println("Poll interval:", cfg.PollInterval, "seconds")

	// Создаём и запускаем агента с интервалами из конфигурации
	a := NewAgent(
		cfg.ServerAddress,
		time.Duration(cfg.PollInterval)*time.Second,
		time.Duration(cfg.ReportInterval)*time.Second,
	)
	a.Run()
}
