package cli

import (
	"strconv"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printServerStatusTable(serverBuildVersion string, serverBuildTime string) {
	t, theme := createTable(1)

	row := []interface{}{
		termenv.String("Server host").Foreground(theme.ColorCyan),
		termenv.String(ServerHost).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Server port").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(ServerPort)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Server version").Foreground(theme.ColorCyan),
		termenv.String(serverBuildVersion).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Server buildtime").Foreground(theme.ColorCyan),
		termenv.String(formatTimestamp(serverBuildTime)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
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

func printServerStatTable(stat *core.Statistics) {
	t, theme := createTable(0)

	row := []interface{}{
		termenv.String("Colonies").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(stat.Colonies)).Foreground(theme.ColorGray),
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
