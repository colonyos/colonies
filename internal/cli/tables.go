package cli

import (
	"errors"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	goprettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/termenv"
)

func printProcessesTable(processes []*core.Process, mode int) {
	style := goprettytable.StyleRounded
	theme, err := table.LoadTheme("solarized-dark")
	CheckError(err)

	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t := table.NewTable(theme, table.TableOptions{
		Columns: []int{1, 2},
		SortBy:  sortCol,
		Style:   style,
	})

	var timeid string
	var timeTitle string

	switch mode {
	case core.WAITING:
		timeid = "submissiontime"
		timeTitle = "SubmssionTime"
	case core.RUNNING:
		timeid = "starttime"
		timeTitle = "StartTime"
	case core.SUCCESS:
		timeid = "endtime"
		timeTitle = "EndTime"
	case core.FAILED:
		timeid = "endtime"
		timeTitle = "EndTime"
	default:
		CheckError(errors.New("Invalid table type"))
	}

	if ShowIDs {
		var cols = []table.Column{
			{ID: "id", Name: "ID", SortIndex: 1},
			{ID: "funcname", Name: "FuncName", SortIndex: 2},
			{ID: "args", Name: "Args", SortIndex: 3},
			{ID: "kwargs", Name: "KwArgs", SortIndex: 4},
			{ID: timeid, Name: timeTitle, SortIndex: 5},
			{ID: "executortype", Name: "ExecutorType", SortIndex: 6},
			{ID: "initiatorname", Name: "InitiatorName", SortIndex: 7},
		}
		t.SetCols(cols)
	} else {
		var cols = []table.Column{
			{ID: "funcname", Name: "FuncName", SortIndex: 1},
			{ID: "args", Name: "Args", SortIndex: 2},
			{ID: "kwargs", Name: "KwArgs", SortIndex: 3},
			{ID: "endtime", Name: "EndTime", SortIndex: 4},
			{ID: "executortype", Name: "ExecutorType", SortIndex: 5},
			{ID: "initiatorname", Name: "InitiatorName", SortIndex: 6},
		}
		t.SetCols(cols)
	}

	for _, process := range processes {
		args, kwArgs := parseArgs(process)
		var timeValue string
		var timeColor termenv.Color
		switch mode {
		case core.WAITING:
			timeValue = process.SubmissionTime.Format(TimeLayout)
			timeColor = theme.ColorBlue
		case core.RUNNING:
			timeValue = process.StartTime.Format(TimeLayout)
			timeColor = theme.ColorCyan
		case core.SUCCESS:
			timeValue = process.EndTime.Format(TimeLayout)
			timeColor = theme.ColorGreen
		case core.FAILED:
			timeValue = process.EndTime.Format(TimeLayout)
			timeColor = theme.ColorRed
		default:
			CheckError(errors.New("Invalid table type"))
		}
		if ShowIDs {
			row := []interface{}{
				termenv.String(process.ID).Foreground(theme.ColorGray),
				termenv.String(process.FunctionSpec.FuncName).Foreground(theme.ColorMagenta),
				termenv.String(args).Foreground(theme.ColorViolet),
				termenv.String(kwArgs).Foreground(theme.ColorViolet),
				termenv.String(timeValue).Foreground(timeColor),
				termenv.String(process.FunctionSpec.Conditions.ExecutorType).Foreground(theme.ColorYellow),
				termenv.String(process.InitiatorName).Foreground(theme.ColorCyan),
			}
			t.AddRow(row)
		} else {
			row := []interface{}{
				termenv.String(process.FunctionSpec.FuncName).Foreground(theme.ColorMagenta),
				termenv.String(args).Foreground(theme.ColorViolet),
				termenv.String(kwArgs).Foreground(theme.ColorViolet),
				termenv.String(timeValue).Foreground(timeColor),
				termenv.String(process.FunctionSpec.Conditions.ExecutorType).Foreground(theme.ColorYellow),
				termenv.String(process.InitiatorName).Foreground(theme.ColorCyan),
			}
			t.AddRow(row)
		}
	}

	t.Render()
}
