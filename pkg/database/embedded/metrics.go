package embedded

import (
	"errors"
	"sort"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *EmbeddedDatabase) SetMetric(metric core.Metric) error {
	metric.GenerateID()
	if err := db.metrics.Put(metric.ID, &metric); err != nil {
		return err
	}
	db.metricsIdx.byExecutor.Add(metric.ID, metric.ColonyName+":"+metric.ExecutorName)
	db.metricsIdx.byColony.Add(metric.ID, metric.ColonyName)
	return nil
}

func (db *EmbeddedDatabase) GetMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) (core.Metric, error) {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()
	existing, ok := db.metrics.Get(m.ID)
	if !ok {
		return core.Metric{}, errors.New("Metric does not exist")
	}
	return *existing, nil
}

// GetMetricsByExecutorName returns only PERIOD_NONE metrics for the executor.
func (db *EmbeddedDatabase) GetMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error) {
	ids := db.metricsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	result := make([]core.Metric, 0, len(ids))
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			if m.Period == core.PERIOD_NONE {
				result = append(result, *m)
			}
		}
	}
	return result, nil
}

// GetAllMetricsByExecutorName returns all metrics for the executor, including period-bucketed ones.
func (db *EmbeddedDatabase) GetAllMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error) {
	ids := db.metricsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	result := make([]core.Metric, 0, len(ids))
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			result = append(result, *m)
		}
	}
	return result, nil
}

// GetMetricsByColonyName returns only PERIOD_NONE metrics for the colony.
func (db *EmbeddedDatabase) GetMetricsByColonyName(colonyName string) ([]core.Metric, error) {
	ids := db.metricsIdx.byColony.Lookup(colonyName)
	result := make([]core.Metric, 0, len(ids))
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			if m.Period == core.PERIOD_NONE {
				result = append(result, *m)
			}
		}
	}
	return result, nil
}

func (db *EmbeddedDatabase) GetMetricHistory(colonyName string, executorName string, key string, period int, from time.Time, to time.Time) ([]core.Metric, error) {
	ids := db.metricsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	var result []core.Metric
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			if m.Key == key && m.Period == period &&
				!m.PeriodStart.Before(from) && !m.PeriodStart.After(to) {
				result = append(result, *m)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PeriodStart.Before(result[j].PeriodStart)
	})
	if result == nil {
		result = make([]core.Metric, 0)
	}
	return result, nil
}

func (db *EmbeddedDatabase) IncrementMetric(colonyName string, executorName string, key string, period int, periodStart time.Time, delta float64) error {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		MetricType:   core.COUNTER,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()

	// Atomic read-modify-write using the store's own lock.
	// This matches PostgreSQL's `UPDATE SET value = value + delta` semantics.
	db.metrics.Lock()
	defer db.metrics.Unlock()

	existing, ok := db.metrics.GetUnlocked(m.ID)
	if !ok {
		m.Value = delta
		if err := db.metrics.PutUnlocked(m.ID, &m); err != nil {
			return err
		}
		db.metricsIdx.byExecutor.Add(m.ID, colonyName+":"+executorName)
		db.metricsIdx.byColony.Add(m.ID, colonyName)
		return nil
	}
	cp := *existing
	cp.Value += delta
	return db.metrics.PutUnlocked(cp.ID, &cp)
}

func (db *EmbeddedDatabase) RemoveMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) error {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()
	_, ok := db.metrics.Get(m.ID)
	if !ok {
		return nil
	}
	db.metricsIdx.byExecutor.Remove(m.ID, colonyName+":"+executorName)
	db.metricsIdx.byColony.Remove(m.ID, colonyName)
	return db.metrics.Delete(m.ID)
}

func (db *EmbeddedDatabase) RemoveAllMetricsByExecutorName(colonyName string, executorName string) error {
	ids := db.metricsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			db.metricsIdx.byExecutor.Remove(id, colonyName+":"+executorName)
			db.metricsIdx.byColony.Remove(id, m.ColonyName)
			db.metrics.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllMetricsByColonyName(colonyName string) error {
	ids := db.metricsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if m, ok := db.metrics.Get(id); ok {
			db.metricsIdx.byExecutor.Remove(id, m.ColonyName+":"+m.ExecutorName)
			db.metricsIdx.byColony.Remove(id, colonyName)
			db.metrics.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllMetrics() error {
	for _, m := range db.metrics.All() {
		db.metrics.Delete(m.ID)
	}
	db.metricsIdx.byExecutor.Clear()
	db.metricsIdx.byColony.Clear()
	return nil
}
