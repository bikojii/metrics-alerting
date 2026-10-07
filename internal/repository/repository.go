package repository

import models "github.com/bikojii/metrics-alerting/internal/model"

type Storage interface {
	SaveMetric(models.Metrics)
	GetMetric(string, string) (models.Metrics, bool)
	ListMetrics() []models.Metrics
}
