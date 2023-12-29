package cli

import (
	"fmt"
	"strconv"

	"github.com/muesli/termenv"
)

func printFunctionTable(statMap map[string]statsEntry) {
	for _, s := range statMap {
		t, theme := createTable(0)
		t.SetTitle("Function: " + s.funcName)

		row := []interface{}{
			termenv.String("ExecutorType").Foreground(theme.ColorBlue),
			termenv.String(s.executorType).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("FunctionName").Foreground(theme.ColorBlue),
			termenv.String(s.funcName).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Call counter").Foreground(theme.ColorBlue),
			termenv.String(strconv.Itoa(s.callsCounter)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MinWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.minWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MaxWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.maxWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("AvgWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.avgWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MinExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.minExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MaxExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.maxExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("AvgExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", s.avgExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}
}
