package cli

import (
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printColonyTable(colonies []*core.Colony) {
	t, theme := createTable(0)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "colonyid", Name: "ColonyId", SortIndex: 2},
	}
	t.SetCols(cols)

	for _, colony := range colonies {
		row := []interface{}{
			termenv.String(colony.Name).Foreground(theme.ColorCyan),
			termenv.String(colony.ID).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printColonyStatTable(stat *core.Statistics) {
	t, theme := createTable(0)

	row := []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Executors").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.Executors)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Waiting processes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.WaitingProcesses)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Running processes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.RunningProcesses)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Successful processes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.SuccessfulProcesses)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Failed processes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.FailedProcesses)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Waiting workflows").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.WaitingWorkflows)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Running workflows").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.RunningWorkflows)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Successful workflows").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.SuccessfulWorkflows)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Failed workflows").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.FailedWorkflows)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}
