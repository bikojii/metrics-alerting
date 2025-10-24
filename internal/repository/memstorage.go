package repository

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
