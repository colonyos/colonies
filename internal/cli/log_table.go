package cli

import (
	"strconv"
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
			termenv.String(log.Message).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		var timeback int64
		timeback = 1000000000 * 1 // 1 seconds
		row = []interface{}{
			termenv.String("Cmd").Foreground(theme.ColorGreen),
			termenv.String("colonies log get --processid " + log.ProcessID + " --since " + strconv.FormatInt(log.Timestamp-timeback, 10)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}

}
