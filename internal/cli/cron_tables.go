package cli

import (
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printCronTable(cron *core.Cron) {
	t, theme := createTable(0)

	t.SetTitle("Cron")

	row := []interface{}{
		termenv.String("CronId").Foreground(theme.ColorCyan),
		termenv.String(cron.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(cron.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(cron.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("InitiatorID").Foreground(theme.ColorCyan),
		termenv.String(cron.InitiatorID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Initiator").Foreground(theme.ColorCyan),
		termenv.String(cron.InitiatorName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Cron Expression").Foreground(theme.ColorCyan),
		termenv.String(cron.CronExpression).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Interval").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(cron.Interval)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Random").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatBool(cron.Random)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("NextRun").Foreground(theme.ColorCyan),
		termenv.String(cron.NextRun.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("LastRun").Foreground(theme.ColorCyan),
		termenv.String(cron.NextRun.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("PrevProcessGraphID").Foreground(theme.ColorCyan),
		termenv.String(cron.PrevProcessGraphID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("WaitForPrevProcessGraph").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatBool(cron.WaitForPrevProcessGraph)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CheckerPeriod").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(cron.CheckerPeriod)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printCronsTable(crons []*core.Cron) {
	t, theme := createTable(0)

	var cols = []table.Column{
		{ID: "cronid", Name: "CronId", SortIndex: 1},
		{ID: "name", Name: "Name", SortIndex: 2},
		{ID: "initiator", Name: "Initiator", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, cron := range crons {
		row := []interface{}{
			termenv.String(cron.ID).Foreground(theme.ColorGray),
			termenv.String(cron.Name).Foreground(theme.ColorCyan),
			termenv.String(cron.InitiatorName).Foreground(theme.ColorViolet),
		}
		t.AddRow(row)
	}

	t.Render()
}
