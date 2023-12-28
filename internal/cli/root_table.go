package cli

import (
	"fmt"
	"strconv"

	"github.com/muesli/termenv"
)

func printConfigTable() {
	t, theme := createTable(0)

	t.SetTitle("Current configuration")

	row := []interface{}{
		termenv.String("ColoniesServer").Foreground(theme.ColorCyan),
		termenv.String(ServerHost + ":" + strconv.Itoa(ServerPort)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("TLS").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%t", UseTLS)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}
