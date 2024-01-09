package cli

import (
	"fmt"
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printExecutorsTable(executors []*core.Executor) {
	t, theme := createTable(2)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "type", Name: "Type", SortIndex: 2},
		{ID: "Location", Name: "Location", SortIndex: 3},
		{ID: "lastheardfrom", Name: "Last Heard From", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, executor := range executors {
		row := []interface{}{
			termenv.String(executor.Name).Foreground(theme.ColorCyan),
			termenv.String(executor.Type).Foreground(theme.ColorViolet),
			termenv.String(executor.Location.Description).Foreground(theme.ColorMagenta),
			termenv.String(executor.LastHeardFromTime.Format(TimeLayout)).Foreground(theme.ColorGreen),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printExecutorTable(client *client.ColoniesClient, executor *core.Executor) {
	state := ""
	switch executor.State {
	case core.PENDING:
		state = "Pending"
	case core.APPROVED:
		state = "Approved"
	case core.REJECTED:
		state = "Rejected"
	default:
		state = "Unknown"
	}

	requireFuncRegStr := "False"
	if executor.RequireFuncReg {
		requireFuncRegStr = "True"
	}

	t, theme := createTable(0)

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(executor.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Id").Foreground(theme.ColorCyan),
		termenv.String(executor.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Type").Foreground(theme.ColorCyan),
		termenv.String(executor.Type).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(executor.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("State").Foreground(theme.ColorCyan),
		termenv.String(state).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("RequireFuncRegistration").Foreground(theme.ColorCyan),
		termenv.String(requireFuncRegStr).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Commission Time").Foreground(theme.ColorCyan),
		termenv.String(executor.CommissionTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Last Heard From").Foreground(theme.ColorCyan),
		termenv.String(executor.LastHeardFromTime.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	t, theme = createTable(0)
	t.SetTitle("Location")

	row = []interface{}{
		termenv.String("Longitude").Foreground(theme.ColorViolet),
		termenv.String(fmt.Sprintf("%f", executor.Location.Long)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Latitude").Foreground(theme.ColorViolet),
		termenv.String(fmt.Sprintf("%f", executor.Location.Lat)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Description").Foreground(theme.ColorViolet),
		termenv.String(executor.Location.Description).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	t, theme = createTable(0)
	t.SetTitle("Hardware")

	row = []interface{}{
		termenv.String("Nodes").Foreground(theme.ColorMagenta),
		termenv.String(strconv.Itoa(executor.Capabilities.Hardware.Nodes)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Model").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.Model).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CPU").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.CPU).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Memory").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.Memory).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Storage").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.Storage).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPU").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.GPU.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUs").Foreground(theme.ColorMagenta),
		termenv.String(strconv.Itoa(executor.Capabilities.Hardware.GPU.Count)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPU/Node").Foreground(theme.ColorMagenta),
		termenv.String(strconv.Itoa(executor.Capabilities.Hardware.GPU.NodeCount)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPU Memory").Foreground(theme.ColorMagenta),
		termenv.String(executor.Capabilities.Hardware.GPU.Memory).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	t, theme = createTable(0)
	t.SetTitle("Software")

	row = []interface{}{
		termenv.String("Name").Foreground(theme.ColorBlue),
		termenv.String(executor.Capabilities.Software.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Type").Foreground(theme.ColorBlue),
		termenv.String(executor.Capabilities.Software.Type).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Version").Foreground(theme.ColorBlue),
		termenv.String(executor.Capabilities.Software.Version).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	functions, err := client.GetFunctionsByExecutor(ColonyName, executor.Name, PrvKey)
	CheckError(err)

	for _, function := range functions {
		t, theme = createTable(0)
		t.SetTitle("Function: " + function.FuncName)

		row = []interface{}{
			termenv.String("FunctionName").Foreground(theme.ColorBlue),
			termenv.String(function.FuncName).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("FunctionId").Foreground(theme.ColorBlue),
			termenv.String(function.FunctionID).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("Call counter").Foreground(theme.ColorBlue),
			termenv.String(strconv.Itoa(function.Counter)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MinWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.MinWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MaxWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.MaxWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("AvgWaitTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.AvgWaitTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MinExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.MinExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("MaxExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.MaxExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("AvgExecTime").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%f s", function.AvgExecTime)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}
}
