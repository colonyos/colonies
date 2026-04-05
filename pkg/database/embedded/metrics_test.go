package embedded

import (
	"sync"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Basic CRUD (PERIOD_NONE, backward compatible) ---

func TestSetMetric(t *testing.T) {
	db := setupTestDB(t)

	metric := core.CreateMetric("test-colony", "test-executor", "gpu_temp", core.GAUGE, 75.5)
	err := db.SetMetric(metric)
	assert.NoError(t, err)

	got, err := db.GetMetric("test-colony", "test-executor", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, "test-colony", got.ColonyName)
	assert.Equal(t, "test-executor", got.ExecutorName)
	assert.Equal(t, "gpu_temp", got.Key)
	assert.Equal(t, core.GAUGE, got.MetricType)
	assert.Equal(t, 75.5, got.Value)
	assert.Equal(t, core.PERIOD_NONE, got.Period)
}

func TestSetMetricOverwrite(t *testing.T) {
	db := setupTestDB(t)

	metric := core.CreateMetric("test-colony", "test-executor", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(metric))

	metric2 := core.CreateMetric("test-colony", "test-executor", "gpu_temp", core.GAUGE, 80.0)
	assert.NoError(t, db.SetMetric(metric2))

	got, err := db.GetMetric("test-colony", "test-executor", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 80.0, got.Value)
}

func TestGetMetricNotFound(t *testing.T) {
	db := setupTestDB(t)
	_, err := db.GetMetric("test-colony", "test-executor", "nonexistent", core.PERIOD_NONE, time.Time{})
	assert.Error(t, err)
}

func TestSetMultipleMetrics(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor1", "gpu_mem", core.GAUGE, 4096.0)
	m3 := core.CreateMetric("colony1", "executor2", "gpu_temp", core.GAUGE, 80.0)
	m4 := core.CreateMetric("colony2", "executor1", "tokens_total", core.COUNTER, 1000.0)

	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))
	assert.NoError(t, db.SetMetric(m3))
	assert.NoError(t, db.SetMetric(m4))

	got, err := db.GetMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 75.5, got.Value)

	got, err = db.GetMetric("colony1", "executor2", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 80.0, got.Value)
}

func TestGetMetricsByExecutorName(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor1", "gpu_mem", core.GAUGE, 4096.0)
	m3 := core.CreateMetric("colony1", "executor2", "gpu_temp", core.GAUGE, 80.0)

	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))
	assert.NoError(t, db.SetMetric(m3))

	metrics, err := db.GetMetricsByExecutorName("colony1", "executor1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	metrics, err = db.GetMetricsByExecutorName("colony1", "executor2")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, 80.0, metrics[0].Value)
}

func TestGetMetricsByExecutorNameEmpty(t *testing.T) {
	db := setupTestDB(t)
	metrics, err := db.GetMetricsByExecutorName("colony1", "nonexistent")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestGetMetricsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor2", "gpu_temp", core.GAUGE, 80.0)
	m3 := core.CreateMetric("colony2", "executor1", "tokens_total", core.COUNTER, 1000.0)

	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))
	assert.NoError(t, db.SetMetric(m3))

	metrics, err := db.GetMetricsByColonyName("colony1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	metrics, err = db.GetMetricsByColonyName("colony2")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestGetMetricsByColonyNameEmpty(t *testing.T) {
	db := setupTestDB(t)
	metrics, err := db.GetMetricsByColonyName("nonexistent")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestIncrementMetricBackwardCompat(t *testing.T) {
	db := setupTestDB(t)

	// PERIOD_NONE with zero time => same behavior as before
	err := db.IncrementMetric("colony1", "executor1", "tokens_total", core.PERIOD_NONE, time.Time{}, 100.0)
	assert.NoError(t, err)

	got, err := db.GetMetric("colony1", "executor1", "tokens_total", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 100.0, got.Value)
	assert.Equal(t, core.COUNTER, got.MetricType)

	err = db.IncrementMetric("colony1", "executor1", "tokens_total", core.PERIOD_NONE, time.Time{}, 50.0)
	assert.NoError(t, err)

	got, err = db.GetMetric("colony1", "executor1", "tokens_total", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 150.0, got.Value)
}

func TestIncrementMetricMultipleTimes(t *testing.T) {
	db := setupTestDB(t)

	for i := 0; i < 100; i++ {
		err := db.IncrementMetric("colony1", "executor1", "requests", core.PERIOD_NONE, time.Time{}, 1.0)
		assert.NoError(t, err)
	}

	got, err := db.GetMetric("colony1", "executor1", "requests", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 100.0, got.Value)
}

func TestRemoveMetric(t *testing.T) {
	db := setupTestDB(t)

	m := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(m))

	err := db.RemoveMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)

	_, err = db.GetMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.Error(t, err)
}

func TestRemoveMetricNotFound(t *testing.T) {
	db := setupTestDB(t)
	err := db.RemoveMetric("colony1", "executor1", "nonexistent", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
}

func TestRemoveMetricDoesNotAffectOthers(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor1", "gpu_mem", core.GAUGE, 4096.0)
	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))

	assert.NoError(t, db.RemoveMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{}))

	got, err := db.GetMetric("colony1", "executor1", "gpu_mem", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 4096.0, got.Value)
}

func TestRemoveAllMetricsByExecutorName(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor1", "gpu_mem", core.GAUGE, 4096.0)
	m3 := core.CreateMetric("colony1", "executor2", "gpu_temp", core.GAUGE, 80.0)
	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))
	assert.NoError(t, db.SetMetric(m3))

	assert.NoError(t, db.RemoveAllMetricsByExecutorName("colony1", "executor1"))

	metrics, err := db.GetMetricsByExecutorName("colony1", "executor1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)

	metrics, err = db.GetMetricsByExecutorName("colony1", "executor2")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestRemoveAllMetricsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony1", "executor2", "gpu_temp", core.GAUGE, 80.0)
	m3 := core.CreateMetric("colony2", "executor1", "tokens_total", core.COUNTER, 1000.0)
	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))
	assert.NoError(t, db.SetMetric(m3))

	assert.NoError(t, db.RemoveAllMetricsByColonyName("colony1"))

	metrics, err := db.GetMetricsByColonyName("colony1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)

	metrics, err = db.GetMetricsByColonyName("colony2")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestRemoveAllMetrics(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	m2 := core.CreateMetric("colony2", "executor2", "tokens_total", core.COUNTER, 1000.0)
	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))

	assert.NoError(t, db.RemoveAllMetrics())

	metrics, err := db.GetMetricsByColonyName("colony1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)

	metrics, err = db.GetMetricsByColonyName("colony2")
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestMetricGaugeAndCounter(t *testing.T) {
	db := setupTestDB(t)

	gauge := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(gauge))
	assert.NoError(t, db.IncrementMetric("colony1", "executor1", "tokens_total", core.PERIOD_NONE, time.Time{}, 500.0))

	metrics, err := db.GetMetricsByExecutorName("colony1", "executor1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	for _, m := range metrics {
		if m.Key == "gpu_temp" {
			assert.Equal(t, core.GAUGE, m.MetricType)
			assert.Equal(t, 75.5, m.Value)
		} else if m.Key == "tokens_total" {
			assert.Equal(t, core.COUNTER, m.MetricType)
			assert.Equal(t, 500.0, m.Value)
		}
	}
}

func TestMetricPersistence(t *testing.T) {
	dir := t.TempDir()

	db := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db.Initialize())

	m := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(m))
	assert.NoError(t, db.IncrementMetric("colony1", "executor1", "tokens", core.PERIOD_NONE, time.Time{}, 100.0))
	db.Close()

	db2 := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db2.Initialize())
	defer db2.Close()

	got, err := db2.GetMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 75.5, got.Value)

	got, err = db2.GetMetric("colony1", "executor1", "tokens", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 100.0, got.Value)

	metrics, err := db2.GetMetricsByExecutorName("colony1", "executor1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
}

func TestMetricSameKeyDifferentExecutors(t *testing.T) {
	db := setupTestDB(t)

	assert.NoError(t, db.IncrementMetric("colony1", "llm-executor-1", "tokens_used", core.PERIOD_NONE, time.Time{}, 500))
	assert.NoError(t, db.IncrementMetric("colony1", "llm-executor-2", "tokens_used", core.PERIOD_NONE, time.Time{}, 300))

	got1, err := db.GetMetric("colony1", "llm-executor-1", "tokens_used", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 500.0, got1.Value)

	got2, err := db.GetMetric("colony1", "llm-executor-2", "tokens_used", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 300.0, got2.Value)
}

func TestMetricSameKeyDifferentColonies(t *testing.T) {
	db := setupTestDB(t)

	m1 := core.CreateMetric("colony1", "executor1", "gpu_temp", core.GAUGE, 75.0)
	m2 := core.CreateMetric("colony2", "executor1", "gpu_temp", core.GAUGE, 82.0)
	assert.NoError(t, db.SetMetric(m1))
	assert.NoError(t, db.SetMetric(m2))

	got1, err := db.GetMetric("colony1", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 75.0, got1.Value)

	got2, err := db.GetMetric("colony2", "executor1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 82.0, got2.Value)
}

// --- Period-based tests ---

func TestIncrementMetricDailyBucket(t *testing.T) {
	db := setupTestDB(t)

	day1 := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	// Increment same day twice => accumulates in one bucket
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day1, 100))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day1, 50))

	got, err := db.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day1)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, got.Value)

	// Next day creates a new bucket
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day2, 200))

	got2, err := db.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day2)
	assert.NoError(t, err)
	assert.Equal(t, 200.0, got2.Value)

	// Day1 is unchanged
	got1, err := db.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day1)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, got1.Value)
}

func TestIncrementMetricMonthlyBucket(t *testing.T) {
	db := setupTestDB(t)

	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_MONTH, jan, 1000))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_MONTH, feb, 2000))

	got1, _ := db.GetMetric("c1", "e1", "tokens", core.PERIOD_MONTH, jan)
	assert.Equal(t, 1000.0, got1.Value)

	got2, _ := db.GetMetric("c1", "e1", "tokens", core.PERIOD_MONTH, feb)
	assert.Equal(t, 2000.0, got2.Value)
}

func TestGetMetricHistory(t *testing.T) {
	db := setupTestDB(t)

	day1 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	day4 := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)
	day5 := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day1, 100))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day2, 200))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day3, 300))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day4, 400))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day5, 500))

	// Query full range
	metrics, err := db.GetMetricHistory("c1", "e1", "tokens", core.PERIOD_DAY, day1, day5)
	assert.NoError(t, err)
	assert.Len(t, metrics, 5)
	// Should be sorted by PeriodStart ascending
	assert.Equal(t, 100.0, metrics[0].Value)
	assert.Equal(t, 500.0, metrics[4].Value)

	// Query partial range
	metrics, err = db.GetMetricHistory("c1", "e1", "tokens", core.PERIOD_DAY, day2, day4)
	assert.NoError(t, err)
	assert.Len(t, metrics, 3)
	assert.Equal(t, 200.0, metrics[0].Value)
	assert.Equal(t, 400.0, metrics[2].Value)
}

func TestGetMetricHistoryEmpty(t *testing.T) {
	db := setupTestDB(t)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	metrics, err := db.GetMetricHistory("c1", "e1", "tokens", core.PERIOD_DAY, from, to)
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestGetMetricHistoryWeeklyMondayStart(t *testing.T) {
	db := setupTestDB(t)

	// 2026-03-30 is Monday, 2026-04-06 is Monday, 2026-04-13 is Monday
	week1 := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	week2 := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	week3 := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, db.IncrementMetric("c1", "e1", "requests", core.PERIOD_WEEK, week1, 10))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "requests", core.PERIOD_WEEK, week2, 20))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "requests", core.PERIOD_WEEK, week3, 30))

	metrics, err := db.GetMetricHistory("c1", "e1", "requests", core.PERIOD_WEEK, week1, week3)
	assert.NoError(t, err)
	assert.Len(t, metrics, 3)
	assert.True(t, metrics[0].PeriodStart.Equal(week1))
	assert.True(t, metrics[1].PeriodStart.Equal(week2))
	assert.True(t, metrics[2].PeriodStart.Equal(week3))
	assert.Equal(t, 10.0, metrics[0].Value)
	assert.Equal(t, 20.0, metrics[1].Value)
	assert.Equal(t, 30.0, metrics[2].Value)
}

func TestGetMetricHistoryDoesNotReturnDifferentPeriod(t *testing.T) {
	db := setupTestDB(t)

	day := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	month := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day, 100))
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_MONTH, month, 5000))

	// Query daily history should not return the monthly metric
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	metrics, err := db.GetMetricHistory("c1", "e1", "tokens", core.PERIOD_DAY, from, to)
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, 100.0, metrics[0].Value)
}

func TestGetMetricsOnlyReturnsPeriodNone(t *testing.T) {
	db := setupTestDB(t)

	// Add PERIOD_NONE metric
	m := core.CreateMetric("c1", "e1", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(m))

	// Add period-bucketed metric for same executor
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC), 100))

	// GetMetricsByExecutorName should only return the PERIOD_NONE metric
	metrics, err := db.GetMetricsByExecutorName("c1", "e1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "gpu_temp", metrics[0].Key)
	assert.Equal(t, core.PERIOD_NONE, metrics[0].Period)

	// Same for GetMetricsByColonyName
	metrics, err = db.GetMetricsByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "gpu_temp", metrics[0].Key)
}

func TestPeriodMetricDoesNotAffectNoneMetric(t *testing.T) {
	db := setupTestDB(t)

	// Set a PERIOD_NONE counter
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_NONE, time.Time{}, 1000))

	// Set a PERIOD_DAY counter with same key
	day := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day, 50))

	// They should be independent
	noneMetric, err := db.GetMetric("c1", "e1", "tokens", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, noneMetric.Value)

	dayMetric, err := db.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, dayMetric.Value)
}

func TestPeriodMetricPersistence(t *testing.T) {
	dir := t.TempDir()

	db := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db.Initialize())

	day := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day, 100))
	db.Close()

	db2 := CreateEmbeddedDatabase(dir)
	assert.NoError(t, db2.Initialize())
	defer db2.Close()

	got, err := db2.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, got.Value)
	assert.Equal(t, core.PERIOD_DAY, got.Period)
	assert.True(t, day.Equal(got.PeriodStart))
}

func TestRemoveAllMetricsIncludesPeriodMetrics(t *testing.T) {
	db := setupTestDB(t)

	m := core.CreateMetric("c1", "e1", "gpu_temp", core.GAUGE, 75.5)
	assert.NoError(t, db.SetMetric(m))

	day := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	assert.NoError(t, db.IncrementMetric("c1", "e1", "tokens", core.PERIOD_DAY, day, 100))

	assert.NoError(t, db.RemoveAllMetricsByExecutorName("c1", "e1"))

	// Both PERIOD_NONE and period metrics should be gone
	_, err := db.GetMetric("c1", "e1", "gpu_temp", core.PERIOD_NONE, time.Time{})
	assert.Error(t, err)

	_, err = db.GetMetric("c1", "e1", "tokens", core.PERIOD_DAY, day)
	assert.Error(t, err)
}

func TestSetGaugeWithPeriod(t *testing.T) {
	db := setupTestDB(t)

	day := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	m := core.Metric{
		ColonyName:   "c1",
		ExecutorName: "e1",
		Key:          "gpu_temp",
		MetricType:   core.GAUGE,
		Value:        75.5,
		Period:       core.PERIOD_DAY,
		PeriodStart:  day,
	}
	m.GenerateID()
	assert.NoError(t, db.SetMetric(m))

	got, err := db.GetMetric("c1", "e1", "gpu_temp", core.PERIOD_DAY, day)
	assert.NoError(t, err)
	assert.Equal(t, 75.5, got.Value)
	assert.Equal(t, core.GAUGE, got.MetricType)

	// Overwrite
	m.Value = 80.0
	assert.NoError(t, db.SetMetric(m))

	got, err = db.GetMetric("c1", "e1", "gpu_temp", core.PERIOD_DAY, day)
	assert.NoError(t, err)
	assert.Equal(t, 80.0, got.Value)
}

// TestConcurrentIncrementMetric verifies that concurrent increments don't lose updates.
// Before the fix, the read-modify-write in IncrementMetric was not atomic,
// causing lost updates under contention (e.g. multiple LLM executors reporting tokens).
func TestConcurrentIncrementMetric(t *testing.T) {
	db := setupTestDB(t)

	const numGoroutines = 50
	const incrementsPerGoroutine = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				err := db.IncrementMetric("c1", "e1", "total_tokens", core.PERIOD_NONE, time.Time{}, 1.0)
				assert.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	got, err := db.GetMetric("c1", "e1", "total_tokens", core.PERIOD_NONE, time.Time{})
	assert.NoError(t, err)
	expected := float64(numGoroutines * incrementsPerGoroutine)
	assert.Equal(t, expected, got.Value, "Lost updates: expected %v but got %v", expected, got.Value)
}
