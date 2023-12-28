package cli

import (
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printGeneratorTable(generator *core.Generator) {
	t, theme := createTable(1)

	row := []interface{}{
		termenv.String("GeneratorId").Foreground(theme.ColorCyan),
		termenv.String(generator.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(generator.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(generator.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Initiator").Foreground(theme.ColorCyan),
		termenv.String(generator.InitiatorName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Trigger").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(generator.Trigger)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Timeout").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(generator.Timeout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Lastrun").Foreground(theme.ColorCyan),
		termenv.String(generator.LastRun.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CheckerPeriod").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(generator.CheckerPeriod)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("QueueSize").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(generator.QueueSize)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printGeneratorsTable(generators []*core.Generator) {
	t, theme := createTable(0)

	var cols = []table.Column{
		{ID: "generatorid", Name: "GeneratorId", SortIndex: 1},
		{ID: "name", Name: "Name", SortIndex: 2},
		{ID: "initiator", Name: "Initiator", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, generator := range generators {
		row := []interface{}{
			termenv.String(generator.ID).Foreground(theme.ColorCyan),
			termenv.String(generator.Name).Foreground(theme.ColorViolet),
			termenv.String(generator.InitiatorName).Foreground(theme.ColorMagenta),
		}
		t.AddRow(row)
	}

	t.Render()

}
