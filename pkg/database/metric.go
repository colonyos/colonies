package database

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

type MetricDatabase interface {
	SetMetric(metric core.Metric) error
	GetMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) (core.Metric, error)
	GetMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error)
	GetAllMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error)
	GetMetricsByColonyName(colonyName string) ([]core.Metric, error)
	GetMetricHistory(colonyName string, executorName string, key string, period int, from time.Time, to time.Time) ([]core.Metric, error)
	IncrementMetric(colonyName string, executorName string, key string, period int, periodStart time.Time, delta float64) error
	RemoveMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) error
	RemoveAllMetricsByExecutorName(colonyName string, executorName string) error
	RemoveAllMetricsByColonyName(colonyName string) error
	RemoveAllMetrics() error
}
