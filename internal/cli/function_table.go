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

		// Show description if available
		if s.description != "" {
			row = []interface{}{
				termenv.String("Description").Foreground(theme.ColorBlue),
				termenv.String(s.description).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		// Show arguments if available
		if len(s.args) > 0 {
			argsStr := ""
			for i, arg := range s.args {
				if i > 0 {
					argsStr += ", "
				}
				reqStr := ""
				if arg.Required {
					reqStr = "*"
				}
				argsStr += fmt.Sprintf("%s%s (%s)", arg.Name, reqStr, arg.Type)
			}
			row = []interface{}{
				termenv.String("Arguments").Foreground(theme.ColorBlue),
				termenv.String(argsStr).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

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
