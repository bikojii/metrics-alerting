package main

import (
	"net/http"
	"testing"
	"time"

	models "github.com/bikojii/metrics-alerting/internal/model"
)

func TestCollectMetrics(t *testing.T) {
	agent := NewAgent("http://localhost:8080", 1*time.Second, 1*time.Second)
	agent.CollectMetrics()
	agent.CollectMetrics()

	if agent.Metrics["PollCount"].Delta == nil {
		t.Fatal("PollCount метрика не собрана")
	}
	if got := *agent.Metrics["PollCount"].Delta; got != 2 {
		t.Fatalf("PollCount = %d, want 2", got)
	}
	if agent.Metrics["RandomValue"].Value == nil {
		t.Error("RandomValue метрика не собрана")
	}
	if len(agent.Metrics) != 18 {
		t.Error("Недостаточно метрик собрано")
	}
}

func TestSendMetric(t *testing.T) {
	testHandler := (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	agent := NewAgent("http://localhost:8080", 1*time.Second, 1*time.Second)
	agent.client.Transport = handlerTransport{handler: testHandler}

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
