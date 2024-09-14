package cli

import (
	"github.com/colonyos/colonies/pkg/build"
	"github.com/muesli/termenv"
)

func printVersionTable() {
	t, theme := createTable(1)

	row := []interface{}{
		termenv.String("CLI version").Foreground(theme.ColorCyan),
		termenv.String(build.BuildVersion).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CLI buildtime").Foreground(theme.ColorCyan),
		termenv.String(formatTimestamp(build.BuildTime)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}
