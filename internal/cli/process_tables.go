package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printProcessTable(process *core.Process) {
	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t, theme := createTable(sortCol)

	t.SetTitle("Process")

	assignedExecutorID := "None"
	if process.AssignedExecutorID != "" {
		assignedExecutorID = process.AssignedExecutorID
	}

	isAssigned := "False"
	if process.IsAssigned {
		isAssigned = "True"
	}

	input := StrArr2Str(IfArr2StringArr(process.Input))
	if len(input) > MaxArgInfoLength {
		input = input[0:MaxArgInfoLength] + "..."
	}

	output := StrArr2Str(IfArr2StringArr(process.Output))
	if len(output) > MaxArgInfoLength {
		output = output[0:MaxArgInfoLength] + "..."
	}

	row := []interface{}{
		termenv.String("Id").Foreground(theme.ColorGreen),
		termenv.String(process.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("IsAssigned").Foreground(theme.ColorGreen),
		termenv.String(isAssigned).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("InitiatorID").Foreground(theme.ColorGreen),
		termenv.String(process.InitiatorID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Initiator").Foreground(theme.ColorGreen),
		termenv.String(process.InitiatorName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("AssignedExecutorID").Foreground(theme.ColorGreen),
		termenv.String(assignedExecutorID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("AssignedExecutorID").Foreground(theme.ColorGreen),
		termenv.String(State2String(process.State)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("PriorityTime").Foreground(theme.ColorGreen),
		termenv.String(strconv.FormatInt(process.PriorityTime, 10)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("SubmissionTime").Foreground(theme.ColorGreen),
		termenv.String(process.SubmissionTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("StartTime").Foreground(theme.ColorGreen),
		termenv.String(process.SubmissionTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("EndTime").Foreground(theme.ColorGreen),
		termenv.String(process.SubmissionTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("WaitDeadline").Foreground(theme.ColorGreen),
		termenv.String(process.WaitDeadline.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ExecDeadline").Foreground(theme.ColorGreen),
		termenv.String(process.ExecDeadline.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("WaitingTime").Foreground(theme.ColorGreen),
		termenv.String(process.WaitingTime().String()).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ProcessingTime").Foreground(theme.ColorGreen),
		termenv.String(process.ProcessingTime().String()).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Retries").Foreground(theme.ColorGreen),
		termenv.String(strconv.Itoa(process.Retries)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Input").Foreground(theme.ColorGreen),
		termenv.String(input).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Output").Foreground(theme.ColorGreen),
		termenv.String(output).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Errors").Foreground(theme.ColorGreen),
		termenv.String(StrArr2Str(process.Errors)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printFunctionSpecTable(funcSpec *core.FunctionSpec) {
	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t, theme := createTable(sortCol)

	t.SetTitle("Function Specification")

	procFunc := funcSpec.FuncName
	if procFunc == "" {
		procFunc = "None"
	}

	procArgs := ""
	for _, procArg := range IfArr2StringArr(funcSpec.Args) {
		procArgs += procArg + " "
	}
	if procArgs == "" {
		procArgs = "None"
	}

	if len(procArgs) > MaxArgInfoLength {
		procArgs = procArgs[0:MaxArgInfoLength] + "..."
	}

	procKwArgs := ""
	for k, procKwArg := range IfMap2StringMap(funcSpec.KwArgs) {
		procKwArgs += k + ":" + procKwArg + " "
	}
	if procKwArgs == "" {
		procKwArgs = "None"
	}

	if len(procKwArgs) > MaxArgInfoLength {
		procKwArgs = procKwArgs[0:MaxArgInfoLength] + "..."
	}

	row := []interface{}{
		termenv.String("Func").Foreground(theme.ColorViolet),
		termenv.String(procFunc).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Args").Foreground(theme.ColorViolet),
		termenv.String(procArgs).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("KwArgs").Foreground(theme.ColorViolet),
		termenv.String(procKwArgs).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("MaxWaitTime").Foreground(theme.ColorViolet),
		termenv.String(strconv.Itoa(funcSpec.MaxWaitTime)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("MaxExecTime").Foreground(theme.ColorViolet),
		termenv.String(strconv.Itoa(funcSpec.MaxExecTime)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("MaxRetries").Foreground(theme.ColorViolet),
		termenv.String(strconv.Itoa(funcSpec.MaxRetries)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Priority").Foreground(theme.ColorViolet),
		termenv.String(strconv.Itoa(funcSpec.Priority)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)
	t.Render()

	row = []interface{}{
		termenv.String("Label").Foreground(theme.ColorViolet),
		termenv.String(funcSpec.Label).Foreground(theme.ColorGray),
	}
	t.AddRow(row)
	t.Render()
}

func printConditionsTable(funcSpec *core.FunctionSpec) {
	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t, theme := createTable(sortCol)

	t.SetTitle("Conditions")

	executorNames := ""
	for _, executorName := range funcSpec.Conditions.ExecutorNames {
		executorNames += executorName + "\n"
	}
	executorNames = strings.TrimSuffix(executorNames, "\n")
	if executorNames == "" {
		executorNames = "None"
	}

	dep := ""
	for _, s := range funcSpec.Conditions.Dependencies {
		dep += s + " "
	}
	if len(dep) > 0 {
		dep = dep[:len(dep)-1]
	}

	row := []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ExecutorNames").Foreground(theme.ColorCyan),
		termenv.String(executorNames).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ExecutorType").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.ExecutorType).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Dependencies").Foreground(theme.ColorCyan),
		termenv.String(dep).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Nodes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(funcSpec.Conditions.Nodes)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CPU").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.CPU).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Memory").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.Memory).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Processes").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(funcSpec.Conditions.Processes)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ProcessesPerNode").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(funcSpec.Conditions.ProcessesPerNode)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Storage").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.Storage).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Walltime").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(int(funcSpec.Conditions.WallTime))).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUName").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.GPU.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUs").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(funcSpec.Conditions.GPU.Count)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUPerNode").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(funcSpec.Conditions.GPU.NodeCount)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUMemory").Foreground(theme.ColorCyan),
		termenv.String(funcSpec.Conditions.GPU.Memory).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printProcessesTable(processes []*core.Process, mode int) {
	t, theme := createTable(0)

	var timeid string
	var timeTitle string

	switch mode {
	case core.WAITING:
		timeid = "submissiontime"
		timeTitle = "Submssion time"
	case core.RUNNING:
		timeid = "starttime"
		timeTitle = "Start time"
	case core.SUCCESS:
		timeid = "endtime"
		timeTitle = "End time"
	case core.FAILED:
		timeid = "endtime"
		timeTitle = "End time"
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
			{ID: "executorname", Name: "Executor name", SortIndex: 6},
			{ID: "executortype", Name: "Executor type", SortIndex: 7},
			{ID: "initiatorname", Name: "Initiator", SortIndex: 8},
			{ID: "label", Name: "Label", SortIndex: 9},
		}
		t.SetCols(cols)
	} else {
		var cols = []table.Column{
			{ID: "funcname", Name: "FuncName", SortIndex: 1},
			{ID: "args", Name: "Args", SortIndex: 2},
			{ID: "kwargs", Name: "KwArgs", SortIndex: 3},
			{ID: timeid, Name: timeTitle, SortIndex: 4},
			{ID: "executorname", Name: "Executor name", SortIndex: 5},
			{ID: "executortype", Name: "Executor type", SortIndex: 6},
			{ID: "initiatorname", Name: "Initiator", SortIndex: 7},
			{ID: "label", Name: "Label", SortIndex: 8},
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

		executorNames := StrArr2Str(process.FunctionSpec.Conditions.ExecutorNames)

		if ShowIDs {
			row := []interface{}{
				termenv.String(process.ID).Foreground(theme.ColorGray),
				termenv.String(process.FunctionSpec.FuncName).Foreground(theme.ColorMagenta),
				termenv.String(args).Foreground(theme.ColorViolet),
				termenv.String(kwArgs).Foreground(theme.ColorViolet),
				termenv.String(timeValue).Foreground(timeColor),
				termenv.String(executorNames).Foreground(theme.ColorYellow),
				termenv.String(process.FunctionSpec.Conditions.ExecutorType).Foreground(theme.ColorYellow),
				termenv.String(process.InitiatorName).Foreground(theme.ColorCyan),
				termenv.String(process.FunctionSpec.Label).Foreground(theme.ColorGreen),
			}
			t.AddRow(row)
		} else {
			row := []interface{}{
				termenv.String(process.FunctionSpec.FuncName).Foreground(theme.ColorMagenta),
				termenv.String(args).Foreground(theme.ColorViolet),
				termenv.String(kwArgs).Foreground(theme.ColorViolet),
				termenv.String(timeValue).Foreground(timeColor),
				termenv.String(executorNames).Foreground(theme.ColorYellow),
				termenv.String(process.FunctionSpec.Conditions.ExecutorType).Foreground(theme.ColorYellow),
				termenv.String(process.InitiatorName).Foreground(theme.ColorCyan),
				termenv.String(process.FunctionSpec.Label).Foreground(theme.ColorGreen),
			}
			t.AddRow(row)
		}
	}

	t.Render()
}
