package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateMetric(t *testing.T) {
	metric := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	assert.Equal(t, "colony1", metric.ColonyName)
	assert.Equal(t, "executor1", metric.ExecutorName)
	assert.Equal(t, "gpu_temp", metric.Key)
	assert.Equal(t, GAUGE, metric.MetricType)
	assert.Equal(t, 75.5, metric.Value)
	assert.Equal(t, PERIOD_NONE, metric.Period)
	assert.True(t, metric.PeriodStart.IsZero())
	assert.NotEmpty(t, metric.ID)
}

func TestMetricGenerateID(t *testing.T) {
	m1 := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	m2 := CreateMetric("colony1", "executor1", "gpu_temp", COUNTER, 100.0)
	// Same colony+executor+key => same ID regardless of type/value
	assert.Equal(t, m1.ID, m2.ID)

	m3 := CreateMetric("colony1", "executor1", "gpu_mem", GAUGE, 4096.0)
	// Different key => different ID
	assert.NotEqual(t, m1.ID, m3.ID)

	m4 := CreateMetric("colony1", "executor2", "gpu_temp", GAUGE, 75.5)
	// Different executor => different ID
	assert.NotEqual(t, m1.ID, m4.ID)

	m5 := CreateMetric("colony2", "executor1", "gpu_temp", GAUGE, 75.5)
	// Different colony => different ID
	assert.NotEqual(t, m1.ID, m5.ID)
}

func TestMetricGenerateIDWithPeriod(t *testing.T) {
	// PERIOD_NONE uses the old hash (colony+executor+key)
	m1 := CreateMetric("colony1", "executor1", "tokens", COUNTER, 100)

	// Period-aware metric with same key gets a different ID
	m2 := Metric{
		ColonyName:   "colony1",
		ExecutorName: "executor1",
		Key:          "tokens",
		MetricType:   COUNTER,
		Period:       PERIOD_DAY,
		PeriodStart:  time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}
	m2.GenerateID()
	assert.NotEqual(t, m1.ID, m2.ID)

	// Same period but different day gets a different ID
	m3 := Metric{
		ColonyName:   "colony1",
		ExecutorName: "executor1",
		Key:          "tokens",
		MetricType:   COUNTER,
		Period:       PERIOD_DAY,
		PeriodStart:  time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
	}
	m3.GenerateID()
	assert.NotEqual(t, m2.ID, m3.ID)

	// Same period and same day gets the same ID
	m4 := Metric{
		ColonyName:   "colony1",
		ExecutorName: "executor1",
		Key:          "tokens",
		MetricType:   COUNTER,
		Period:       PERIOD_DAY,
		PeriodStart:  time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}
	m4.GenerateID()
	assert.Equal(t, m2.ID, m4.ID)
}

func TestMetricEquals(t *testing.T) {
	m1 := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	m2 := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	assert.True(t, m1.Equals(m2))

	m3 := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 80.0)
	assert.False(t, m1.Equals(m3))

	m4 := CreateMetric("colony1", "executor1", "gpu_temp", COUNTER, 75.5)
	assert.False(t, m1.Equals(m4))
}

func TestMetricEqualsWithPeriod(t *testing.T) {
	ps := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	m1 := Metric{
		ColonyName: "c1", ExecutorName: "e1", Key: "k",
		MetricType: COUNTER, Value: 100, Period: PERIOD_DAY, PeriodStart: ps,
	}
	m1.GenerateID()

	m2 := m1
	assert.True(t, m1.Equals(m2))

	// Different period => not equal
	m3 := m1
	m3.Period = PERIOD_MONTH
	m3.GenerateID()
	assert.False(t, m1.Equals(m3))
}

func TestMetricToJSON(t *testing.T) {
	metric := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	jsonStr, err := metric.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	parsed, err := ConvertJSONToMetric(jsonStr)
	assert.NoError(t, err)
	assert.True(t, metric.Equals(parsed))
}

func TestMetricToJSONWithPeriod(t *testing.T) {
	ps := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	metric := Metric{
		ColonyName: "c1", ExecutorName: "e1", Key: "tokens",
		MetricType: COUNTER, Value: 500, Period: PERIOD_DAY, PeriodStart: ps,
	}
	metric.GenerateID()

	jsonStr, err := metric.ToJSON()
	assert.NoError(t, err)

	parsed, err := ConvertJSONToMetric(jsonStr)
	assert.NoError(t, err)
	assert.True(t, metric.Equals(parsed))
	assert.Equal(t, PERIOD_DAY, parsed.Period)
	assert.True(t, ps.Equal(parsed.PeriodStart))
}

func TestConvertJSONToMetricInvalid(t *testing.T) {
	_, err := ConvertJSONToMetric("invalid json")
	assert.Error(t, err)
}

func TestConvertMetricArrayToJSON(t *testing.T) {
	m1 := CreateMetric("colony1", "executor1", "gpu_temp", GAUGE, 75.5)
	m2 := CreateMetric("colony1", "executor1", "gpu_mem", GAUGE, 4096.0)

	jsonStr, err := ConvertMetricArrayToJSON([]Metric{m1, m2})
	assert.NoError(t, err)

	parsed, err := ConvertJSONToMetricArray(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, parsed, 2)
}

func TestConvertMetricArrayToJSONEmpty(t *testing.T) {
	jsonStr, err := ConvertMetricArrayToJSON([]Metric{})
	assert.NoError(t, err)

	parsed, err := ConvertJSONToMetricArray(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, parsed, 0)
}

func TestMetricConstants(t *testing.T) {
	assert.Equal(t, 0, GAUGE)
	assert.Equal(t, 1, COUNTER)
}

func TestPeriodConstants(t *testing.T) {
	assert.Equal(t, 0, PERIOD_NONE)
	assert.Equal(t, 1, PERIOD_DAY)
	assert.Equal(t, 2, PERIOD_WEEK)
	assert.Equal(t, 3, PERIOD_MONTH)
}

func TestCalculatePeriodStartNone(t *testing.T) {
	now := time.Date(2026, 4, 5, 14, 30, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_NONE, now)
	assert.True(t, ps.IsZero())
}

func TestCalculatePeriodStartDay(t *testing.T) {
	now := time.Date(2026, 4, 5, 14, 30, 45, 123, time.UTC)
	ps := CalculatePeriodStart(PERIOD_DAY, now)
	expected := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartDayMidnight(t *testing.T) {
	now := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_DAY, now)
	assert.True(t, ps.Equal(now))
}

func TestCalculatePeriodStartWeekMonday(t *testing.T) {
	// 2026-04-06 is a Monday
	monday := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_WEEK, monday)
	expected := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartWeekWednesday(t *testing.T) {
	// 2026-04-08 is a Wednesday, week starts on Monday 2026-04-06
	wednesday := time.Date(2026, 4, 8, 15, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_WEEK, wednesday)
	expected := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartWeekSunday(t *testing.T) {
	// 2026-04-05 is a Sunday, week started on Monday 2026-03-30
	sunday := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_WEEK, sunday)
	expected := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartWeekSaturday(t *testing.T) {
	// 2026-04-04 is a Saturday, week started on Monday 2026-03-30
	saturday := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_WEEK, saturday)
	expected := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartMonth(t *testing.T) {
	now := time.Date(2026, 4, 15, 14, 30, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_MONTH, now)
	expected := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}

func TestCalculatePeriodStartMonthFirstDay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ps := CalculatePeriodStart(PERIOD_MONTH, now)
	assert.True(t, ps.Equal(now))
}

func TestCalculatePeriodStartNonUTC(t *testing.T) {
	// Input in non-UTC should be converted to UTC
	loc := time.FixedZone("CET", 3600)
	now := time.Date(2026, 4, 5, 1, 30, 0, 0, loc) // 00:30 UTC
	ps := CalculatePeriodStart(PERIOD_DAY, now)
	expected := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	assert.True(t, ps.Equal(expected))
}
