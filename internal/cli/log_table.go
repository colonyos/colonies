package cli

import (
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printLogTable(logs []*core.Log) {
	for _, log := range logs {
		t, theme := createTable(0)
		row := []interface{}{
			termenv.String("Timestamp").Foreground(theme.ColorCyan),
			termenv.String(time.Unix(0, log.Timestamp).Format(TimeLayout)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("ExecutorName").Foreground(theme.ColorViolet),
			termenv.String(log.ExecutorName).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("ProcessID").Foreground(theme.ColorBlue),
			termenv.String(log.ProcessID).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Text").Foreground(theme.ColorMagenta),
			termenv.String(insertNewLines(strings.TrimSpace(log.Message), 64)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}
}
