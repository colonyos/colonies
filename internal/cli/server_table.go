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
			switch backend.Type {
			case "http":
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

			case "grpc":
				row = []interface{}{
					termenv.String("  gRPC Port").Foreground(theme.ColorCyan),
					termenv.String(strconv.Itoa(backend.Port)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)

				row = []interface{}{
					termenv.String("  gRPC Insecure").Foreground(theme.ColorCyan),
					termenv.String(strconv.FormatBool(backend.Insecure)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)

			case "libp2p":
				row = []interface{}{
					termenv.String("  LibP2P Port").Foreground(theme.ColorCyan),
					termenv.String(strconv.Itoa(backend.Port)).Foreground(theme.ColorGray),
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
		termenv.String("Client Backends").Foreground(theme.ColorYellow).Bold(),
		termenv.String(getEnvWithDefault("COLONIES_CLIENT_BACKENDS", "http")).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	// HTTP Client Configuration
	if strings.Contains(getEnvWithDefault("COLONIES_CLIENT_BACKENDS", "http"), "http") {
		row = []interface{}{
			termenv.String("  HTTP Host").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_HTTP_HOST", "localhost")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("  HTTP Port").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_HTTP_PORT", "50080")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("  HTTP Insecure").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_HTTP_INSECURE", "false")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	// gRPC Client Configuration
	if strings.Contains(getEnvWithDefault("COLONIES_CLIENT_BACKENDS", ""), "grpc") {
		row = []interface{}{
			termenv.String("  gRPC Host").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_GRPC_HOST", "localhost")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("  gRPC Port").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_GRPC_PORT", "50051")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		row = []interface{}{
			termenv.String("  gRPC Insecure").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_GRPC_INSECURE", "false")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	// LibP2P Client Configuration
	if strings.Contains(getEnvWithDefault("COLONIES_CLIENT_BACKENDS", ""), "libp2p") {
		row = []interface{}{
			termenv.String("  LibP2P Host").Foreground(theme.ColorCyan),
			termenv.String(getEnvWithDefault("COLONIES_CLIENT_LIBP2P_HOST", "dht")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

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
