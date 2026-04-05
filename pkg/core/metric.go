package core

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
)

const (
	GAUGE   int = 0
	COUNTER     = 1
)

const (
	PERIOD_NONE  int = 0
	PERIOD_DAY       = 1
	PERIOD_WEEK      = 2 // Monday start (ISO 8601)
	PERIOD_MONTH     = 3
)

type Metric struct {
	ID           string    `json:"metricid"`
	ColonyName   string    `json:"colonyname"`
	ExecutorName string    `json:"executorname"`
	Key          string    `json:"key"`
	MetricType   int       `json:"metrictype"`
	Value        float64   `json:"value"`
	Period       int       `json:"period"`
	PeriodStart  time.Time `json:"periodstart"`
}

func CreateMetric(colonyName string, executorName string, key string, metricType int, value float64) Metric {
	metric := Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		MetricType:   metricType,
		Value:        value,
		Period:       PERIOD_NONE,
	}
	metric.GenerateID()
	return metric
}

func CalculatePeriodStart(period int, now time.Time) time.Time {
	now = now.UTC()
	switch period {
	case PERIOD_DAY:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case PERIOD_WEEK:
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -int(weekday-time.Monday))
		return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
	case PERIOD_MONTH:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return time.Time{}
	}
}

func (metric *Metric) GenerateID() {
	crypto := crypto.CreateCrypto()
	if metric.Period != PERIOD_NONE {
		metric.ID = crypto.GenerateHash(metric.ColonyName + metric.ExecutorName + metric.Key + strconv.Itoa(metric.Period) + metric.PeriodStart.Format("2006-01-02"))
	} else {
		metric.ID = crypto.GenerateHash(metric.ColonyName + metric.ExecutorName + metric.Key)
	}
}

func (metric *Metric) Equals(metric2 Metric) bool {
	if metric.ID == metric2.ID &&
		metric.ColonyName == metric2.ColonyName &&
		metric.ExecutorName == metric2.ExecutorName &&
		metric.Key == metric2.Key &&
		metric.MetricType == metric2.MetricType &&
		metric.Value == metric2.Value &&
		metric.Period == metric2.Period &&
		metric.PeriodStart.Equal(metric2.PeriodStart) {
		return true
	}
	return false
}

func (metric *Metric) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(metric)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func ConvertJSONToMetric(jsonString string) (Metric, error) {
	var metric Metric
	err := json.Unmarshal([]byte(jsonString), &metric)
	if err != nil {
		return metric, err
	}
	return metric, nil
}

func ConvertJSONToMetricArray(jsonString string) ([]Metric, error) {
	var metrics []Metric
	err := json.Unmarshal([]byte(jsonString), &metrics)
	if err != nil {
		return metrics, err
	}
	return metrics, nil
}

func ConvertMetricArrayToJSON(metrics []Metric) (string, error) {
	jsonBytes, err := json.Marshal(metrics)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
