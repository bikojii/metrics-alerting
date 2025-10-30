package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/admin/metrics-alerting/internal/handler"
	"github.com/admin/metrics-alerting/internal/model"
	"github.com/admin/metrics-alerting/internal/repository"
)

func TestUpdateHandlerGauge(t *testing.T) {
	store := repository.NewMemStorage()
	h := handler.UpdateHandler(store)

	// Создаём HTTP-запрос для обновления Gauge метрики
	req := httptest.NewRequest("POST", "/update/gauge/RandomValue/3.14", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", resp.StatusCode)
	}

	// Проверяем через GetMetric
	metric, ok := store.GetMetric("RandomValue", model.Gauge)
	if !ok {
		t.Fatal("Метрика RandomValue не найдена в хранилище")
	}
	if metric.Value == nil || *metric.Value != 3.14 {
		t.Errorf("Ожидалось значение 3.14, получили %v", metric.Value)
	}
}

func TestUpdateHandlerCounter(t *testing.T) {
	store := repository.NewMemStorage()
	h := handler.UpdateHandler(store)

	// Отправляем Counter метрику
	req := httptest.NewRequest("POST", "/update/counter/PollCount/7", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", resp.StatusCode)
	}

	metric, ok := store.GetMetric("PollCount", model.Counter)
	if !ok {
		t.Fatal("Метрика PollCount не найдена")
	}
	if metric.Delta == nil || *metric.Delta != 7 {
		t.Errorf("Ожидалось значение 7, получили %v", metric.Delta)
	}
}
