package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printWorkflowTable(graphs []*core.ProcessGraph, mode int) {
	var sortCol int
	if ShowIDs {
		sortCol = 2
	} else {
		sortCol = 2
	}

	t, theme := createTable(sortCol)

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
	case core.CANCELLED:
		timeid = "endtime"
		timeTitle = "EndTime"
	default:
		CheckError(errors.New("Invalid table type"))
	}

	var cols = []table.Column{
		{ID: "graphid", Name: "WorkflowId", SortIndex: 1},
		{ID: timeid, Name: timeTitle, SortIndex: 2},
		{ID: "initiator", Name: "Initiator", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, graph := range graphs {
		var timeValue string
		var timeColor termenv.Color
		switch mode {
		case core.WAITING:
			timeValue = graph.SubmissionTime.Format(TimeLayout)
			timeColor = theme.ColorBlue
		case core.RUNNING:
			timeValue = graph.StartTime.Format(TimeLayout)
			timeColor = theme.ColorCyan
		case core.SUCCESS:
			timeValue = graph.EndTime.Format(TimeLayout)
			timeColor = theme.ColorGreen
		case core.FAILED:
			timeValue = graph.EndTime.Format(TimeLayout)
			timeColor = theme.ColorRed
		case core.CANCELLED:
			timeValue = graph.EndTime.Format(TimeLayout)
			timeColor = theme.ColorYellow
		default:
			CheckError(errors.New("Invalid table type"))
		}
		row := []interface{}{
			termenv.String(graph.ID).Foreground(theme.ColorGray),
			termenv.String(timeValue).Foreground(timeColor),
			termenv.String(graph.InitiatorName).Foreground(theme.ColorViolet),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printGraf(client *client.ColoniesClient, graph *core.ProcessGraph) {
	t, theme := createTable(1)

	t.SetTitle("Workflow")

	row := []interface{}{
		termenv.String("WorkflowId").Foreground(theme.ColorViolet),
		termenv.String(graph.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("InitiatorId").Foreground(theme.ColorViolet),
		termenv.String(graph.InitiatorID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("InitiatorName").Foreground(theme.ColorViolet),
		termenv.String(graph.InitiatorName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("State").Foreground(theme.ColorViolet),
		termenv.String(State2String(graph.State)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("SubmissionTime").Foreground(theme.ColorViolet),
		termenv.String(graph.SubmissionTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("StartTime").Foreground(theme.ColorViolet),
		termenv.String(graph.StartTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("EndTime").Foreground(theme.ColorViolet),
		termenv.String(graph.EndTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	fmt.Println("\nProcesses:")
	for _, processID := range graph.ProcessIDs {
		t, theme := createTable(1)
		process, err := client.GetProcess(processID, PrvKey)
		CheckError(err)

		f := process.FunctionSpec.FuncName
		if f == "" {
			f = "None"
		}

		procArgs := ""
		for _, procArg := range IfArr2StringArr(process.FunctionSpec.Args) {
			procArgs += procArg + " "
		}
		if procArgs == "" {
			procArgs = "None"
		}

		procKwArgs := ""
		for k, procKwArg := range IfMap2StringMap(process.FunctionSpec.KwArgs) {
			procKwArgs += k + ":" + procKwArg + " "
		}
		if procKwArgs == "" {
			procKwArgs = "None"
		}

		dependencies := ""
		for _, dependency := range process.FunctionSpec.Conditions.Dependencies {
			dependencies += dependency + " "
		}
		if dependencies == "" {
			dependencies = "None"
		}

		row = []interface{}{
			termenv.String("NodeName").Foreground(theme.ColorCyan),
			termenv.String(process.FunctionSpec.NodeName).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("InitiatorId").Foreground(theme.ColorCyan),
			termenv.String(process.InitiatorID).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Initiator").Foreground(theme.ColorCyan),
			termenv.String(process.InitiatorName).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("ProcessId").Foreground(theme.ColorCyan),
			termenv.String(process.ID).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("ExecutorType").Foreground(theme.ColorCyan),
			termenv.String(process.FunctionSpec.Conditions.ExecutorType).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("FuncName").Foreground(theme.ColorCyan),
			termenv.String(f).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Args").Foreground(theme.ColorCyan),
			termenv.String(procArgs).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("KwArgs").Foreground(theme.ColorCyan),
			termenv.String(procKwArgs).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("State").Foreground(theme.ColorCyan),
			termenv.String(State2String(process.State)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("WaitingForParents").Foreground(theme.ColorCyan),
			termenv.String(strconv.FormatBool(process.WaitForParents)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Dependencies").Foreground(theme.ColorCyan),
			termenv.String(dependencies).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}
}
