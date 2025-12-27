package cli

import (
	"fmt"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printLocationTable(location *core.Location) {
	t, theme := createTable(1)

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(location.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("LocationID").Foreground(theme.ColorCyan),
		termenv.String(location.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(location.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Description").Foreground(theme.ColorCyan),
		termenv.String(location.Description).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Longitude").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%f", location.Long)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Latitude").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%f", location.Lat)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printLocationsTable(locations []*core.Location) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "Name", Name: "Name", SortIndex: 1},
		{ID: "Description", Name: "Description", SortIndex: 2},
		{ID: "Long", Name: "Long", SortIndex: 3},
		{ID: "Lat", Name: "Lat", SortIndex: 4},
	}
	t.SetCols(cols)

	for _, location := range locations {
		row := []interface{}{
			termenv.String(location.Name).Foreground(theme.ColorCyan),
			termenv.String(location.Description).Foreground(theme.ColorViolet),
			termenv.String(fmt.Sprintf("%.4f", location.Long)).Foreground(theme.ColorMagenta),
			termenv.String(fmt.Sprintf("%.4f", location.Lat)).Foreground(theme.ColorMagenta),
		}
		t.AddRow(row)
	}

	t.Render()
}
