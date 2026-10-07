package repository

import (
	"sort"
	"sync"

	models "github.com/bikojii/metrics-alerting/internal/model"
)

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{gauges: make(map[string]float64), counters: make(map[string]int64)}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.gauges[name]
	return v, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.counters[name]
	return v, ok
}

func (m *MemStorage) SaveMetric(metric models.Metrics) {
	switch metric.MType {
	case models.Gauge:
		if metric.Value != nil {
			m.UpdateGauge(metric.ID, *metric.Value)
		}
	case models.Counter:
		if metric.Delta != nil {
			m.UpdateCounter(metric.ID, *metric.Delta)
		}
	}
}

func (m *MemStorage) GetMetric(id, mtype string) (models.Metrics, bool) {
	switch mtype {
	case models.Gauge:
		if v, ok := m.GetGauge(id); ok {
			return models.Metrics{ID: id, MType: mtype, Value: &v}, true
		}
	case models.Counter:
		if v, ok := m.GetCounter(id); ok {
			return models.Metrics{ID: id, MType: mtype, Delta: &v}, true
		}
	}
	return models.Metrics{}, false
}

func (m *MemStorage) ListMetrics() []models.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := make([]models.Metrics, 0, len(m.gauges)+len(m.counters))
	for name, value := range m.gauges {
		v := value
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &v})
	}
	for name, delta := range m.counters {
		d := delta
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Counter, Delta: &d})
	}
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].ID == metrics[j].ID {
			return metrics[i].MType < metrics[j].MType
		}
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}
