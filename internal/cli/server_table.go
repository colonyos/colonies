package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printServerStatusTable(serverInfo *core.ServerInfo) {
	t, theme := createTable(1)

	// Basic server info
	row := []interface{}{
		termenv.String("Server version").Foreground(theme.ColorCyan),
		termenv.String(serverInfo.BuildVersion).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Server buildtime").Foreground(theme.ColorCyan),
		termenv.String(formatTimestamp(serverInfo.BuildTime)).Foreground(theme.ColorGray),
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

	// Server Backend configuration section
	if len(serverInfo.Backends) > 0 {
		row = []interface{}{
			termenv.String("").Foreground(theme.ColorCyan),
			termenv.String("").Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		// Show backend types
		backendTypes := make([]string, 0)
		for _, backend := range serverInfo.Backends {
			backendTypes = append(backendTypes, backend.Type)
		}
		row = []interface{}{
			termenv.String("Server Backends").Foreground(theme.ColorYellow).Bold(),
			termenv.String(strings.Join(backendTypes, ", ")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		// Show details for each backend
		for _, backend := range serverInfo.Backends {
			if backend.Type == "http" {
				row = []interface{}{
					termenv.String("  HTTP Host").Foreground(theme.ColorCyan),
					termenv.String(backend.Host).Foreground(theme.ColorGray),
				}
				t.AddRow(row)

				row = []interface{}{
					termenv.String("  HTTP Port").Foreground(theme.ColorCyan),
					termenv.String(strconv.Itoa(backend.Port)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)

				row = []interface{}{
					termenv.String("  HTTP TLS").Foreground(theme.ColorCyan),
					termenv.String(strconv.FormatBool(backend.TLS)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}
		}
	}

	// Client configuration section (from local environment)
	row = []interface{}{
		termenv.String("").Foreground(theme.ColorCyan),
		termenv.String("").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Client Configuration").Foreground(theme.ColorYellow).Bold(),
		termenv.String("").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("  HTTP Host").Foreground(theme.ColorCyan),
		termenv.String(getEnvWithDefault("COLONIES_SERVER_HOST", "localhost")).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("  HTTP Port").Foreground(theme.ColorCyan),
		termenv.String(getEnvWithDefault("COLONIES_SERVER_PORT", "50080")).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("  TLS").Foreground(theme.ColorCyan),
		termenv.String(getEnvWithDefault("COLONIES_TLS", "false")).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
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
