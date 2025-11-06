package repository

import models "github.com/bikojii/metrics-alerting/internal/model"

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.Gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.Counters[name] += delta
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := m.Gauges[name]
	return v, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := m.Counters[name]
	return v, ok
}

func (m *MemStorage) SaveMetric(metric models.Metrics) {
	switch metric.MType {
	case models.Gauge:
		if metric.Value != nil {
			m.Gauges[metric.ID] = *metric.Value
		}
	case models.Counter:
		if metric.Delta != nil {
			m.Counters[metric.ID] += *metric.Delta
		}
	}
}

func (m *MemStorage) GetMetric(id string, mtype string) (models.Metrics, bool) {
	switch mtype {
	case models.Gauge:
		if v, ok := m.Gauges[id]; ok {
			return models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &v,
			}, true
		}

	case models.Counter:
		if v, ok := m.Counters[id]; ok {
			return models.Metrics{
				ID:    id,
				MType: models.Counter,
				Delta: &v,
			}, true
		}
	}
	return models.Metrics{}, false
}
