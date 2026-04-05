package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func metricTypeStr(metricType int) string {
	if metricType == core.COUNTER {
		return "counter"
	}
	return "gauge"
}

func formatMetricValue(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return fmt.Sprintf("%.2f", value)
}

func periodStr(period int) string {
	switch period {
	case core.PERIOD_DAY:
		return "day"
	case core.PERIOD_WEEK:
		return "week"
	case core.PERIOD_MONTH:
		return "month"
	default:
		return "none"
	}
}

func printMetricKeysTable(metrics []core.Metric) {
	type keyInfo struct {
		metricType int
		periods    map[int]bool
	}

	grouped := make(map[string]*keyInfo)
	for _, m := range metrics {
		ki, ok := grouped[m.Key]
		if !ok {
			ki = &keyInfo{metricType: m.MetricType, periods: make(map[int]bool)}
			grouped[m.Key] = ki
		}
		ki.periods[m.Period] = true
	}

	// Sort keys for stable output
	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t, theme := createTable(1)

	cols := []table.Column{
		{ID: "key", Name: "Key", SortIndex: 1},
		{ID: "type", Name: "Type", SortIndex: 2},
		{ID: "periods", Name: "Periods", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, key := range keys {
		ki := grouped[key]
		var periodNames []string
		for p := range ki.periods {
			periodNames = append(periodNames, periodStr(p))
		}
		sort.Strings(periodNames)

		row := []interface{}{
			termenv.String(key).Foreground(theme.ColorCyan),
			termenv.String(metricTypeStr(ki.metricType)).Foreground(theme.ColorViolet),
			termenv.String(strings.Join(periodNames, ", ")).Foreground(theme.ColorGreen),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printMetricsListTable(metrics []core.Metric) {
	t, theme := createTable(1)

	cols := []table.Column{
		{ID: "key", Name: "Key", SortIndex: 1},
		{ID: "value", Name: "Value", SortIndex: 2},
		{ID: "type", Name: "Type", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, metric := range metrics {
		row := []interface{}{
			termenv.String(metric.Key).Foreground(theme.ColorCyan),
			termenv.String(formatMetricValue(metric.Value)).Foreground(theme.ColorGreen),
			termenv.String(metricTypeStr(metric.MetricType)).Foreground(theme.ColorViolet),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printMetricDetailTable(metric core.Metric) {
	t, theme := createTable(0)

	row := []interface{}{
		termenv.String("Executor").Foreground(theme.ColorCyan),
		termenv.String(metric.ExecutorName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Key").Foreground(theme.ColorCyan),
		termenv.String(metric.Key).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Value").Foreground(theme.ColorCyan),
		termenv.String(formatMetricValue(metric.Value)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Type").Foreground(theme.ColorCyan),
		termenv.String(metricTypeStr(metric.MetricType)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

type executorMetricCount struct {
	Name  string
	Count int
}

func printExecutorMetricCountTable(results []executorMetricCount) {
	t, theme := createTable(1)

	cols := []table.Column{
		{ID: "executor", Name: "Executor", SortIndex: 1},
		{ID: "metrics", Name: "Metrics", SortIndex: 2},
	}
	t.SetCols(cols)

	for _, r := range results {
		row := []interface{}{
			termenv.String(r.Name).Foreground(theme.ColorCyan),
			termenv.String(strconv.Itoa(r.Count)).Foreground(theme.ColorGreen),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printMetricHistoryTable(metrics []core.Metric) {
	t, theme := createTable(1)

	cols := []table.Column{
		{ID: "periodstart", Name: "Period Start", SortIndex: 1},
		{ID: "value", Name: "Value", SortIndex: 2},
	}
	t.SetCols(cols)

	for _, metric := range metrics {
		row := []interface{}{
			termenv.String(metric.PeriodStart.Format("2006-01-02")).Foreground(theme.ColorCyan),
			termenv.String(formatMetricValue(metric.Value)).Foreground(theme.ColorGreen),
		}
		t.AddRow(row)
	}

	t.Render()
}
