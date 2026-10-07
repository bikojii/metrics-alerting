package repository

import (
	"sync"
	"testing"

	models "github.com/bikojii/metrics-alerting/internal/model"
)

func TestConcurrentStorage(t *testing.T) {
	store := NewMemStorage()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				delta, value := int64(1), float64(n)
				store.SaveMetric(models.Metrics{ID: "polls", MType: models.Counter, Delta: &delta})
				store.UpdateGauge("alloc", value)
				store.GetMetric("polls", models.Counter)
				store.ListMetrics()
			}
		}()
	}
	wg.Wait()
	got, ok := store.GetCounter("polls")
	if !ok || got != 2000 {
		t.Fatalf("counter = %d, found = %v", got, ok)
	}
	snapshot := store.ListMetrics()
	*snapshot[0].Value = -1
	if value, _ := store.GetGauge("alloc"); value == -1 {
		t.Fatal("snapshot modifies storage")
	}
}
