package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "github.com/bikojii/metrics-alerting/internal/model"
)

// Тестируем сбор метрик
func TestCollectMetrics(t *testing.T) {
	agent := NewAgent("http://localhost:8080", 1*time.Second, 1*time.Second)
	agent.CollectMetrics()

	if agent.Metrics["PollCount"].Delta == nil {
		t.Error("PollCount метрика не собрана")
	}
	if agent.Metrics["RandomValue"].Value == nil {
		t.Error("RandomValue метрика не собрана")
	}
	if len(agent.Metrics) < 3 { // 2 кастомных + runtime метрики
		t.Error("Недостаточно метрик собрано")
	}
}

// Тестируем отправку метрик
func TestSendMetric(t *testing.T) {
	// Мок-сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	agent := NewAgent(ts.URL, 1*time.Second, 1*time.Second)

	m := &models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: new(float64),
	}
	*m.Value = 42.0

	if err := agent.SendMetric(m); err != nil {
		t.Errorf("SendMetric вернул ошибку: %v", err)
	}
}
