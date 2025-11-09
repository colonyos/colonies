package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

// formatDuration formats a time to show how long ago it was
func formatDuration(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

func printNodesTable(nodes []*core.Node, executors []*core.Executor) {
	t, theme := createTable(2)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "location", Name: "Location", SortIndex: 2},
		{ID: "platform", Name: "Platform", SortIndex: 3},
		{ID: "arch", Name: "Arch", SortIndex: 4},
		{ID: "state", Name: "State", SortIndex: 5},
		{ID: "executors", Name: "Executors", SortIndex: 6},
		{ID: "lastseen", Name: "Last Seen", SortIndex: 7},
	}
	t.SetCols(cols)

	for _, node := range nodes {
		// Count executors for this node
		executorCount := 0
		for _, executor := range executors {
			if executor.NodeID == node.ID {
				executorCount++
			}
		}

		platformArch := fmt.Sprintf("%s/%s", node.Platform, node.Architecture)

		// Calculate time since last seen
		timeSinceLastSeen := formatDuration(node.LastSeen)

		// Determine status color based on activity
		stateColor := theme.ColorGreen
		if executorCount == 0 {
			stateColor = theme.ColorYellow // No executors - warning
		}

		row := []interface{}{
			termenv.String(node.Name).Foreground(theme.ColorCyan),
			termenv.String(node.Location).Foreground(theme.ColorMagenta),
			termenv.String(platformArch).Foreground(theme.ColorViolet),
			termenv.String(node.Architecture).Foreground(theme.ColorYellow),
			termenv.String(node.State).Foreground(stateColor),
			termenv.String(fmt.Sprintf("%d", executorCount)).Foreground(theme.ColorBlue),
			termenv.String(timeSinceLastSeen).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printNodeTable(node *core.Node, executors []*core.Executor) {
	t, theme := createTable(0)

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(node.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Id").Foreground(theme.ColorCyan),
		termenv.String(node.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(node.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Location").Foreground(theme.ColorCyan),
		termenv.String(node.Location).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Platform").Foreground(theme.ColorCyan),
		termenv.String(node.Platform).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Architecture").Foreground(theme.ColorCyan),
		termenv.String(node.Architecture).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("CPU Cores").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(node.CPU)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Memory (MB)").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatInt(node.Memory, 10)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("GPUs").Foreground(theme.ColorCyan),
		termenv.String(strconv.Itoa(node.GPU)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("State").Foreground(theme.ColorCyan),
		termenv.String(node.State).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Created").Foreground(theme.ColorCyan),
		termenv.String(node.Created.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Last Seen").Foreground(theme.ColorCyan),
		termenv.String(node.LastSeen.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	// Print capabilities
	if len(node.Capabilities) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Capabilities")

		row = []interface{}{
			termenv.String("Capabilities").Foreground(theme.ColorViolet),
			termenv.String(strings.Join(node.Capabilities, ", ")).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()
	}

	// Print labels
	if len(node.Labels) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Labels")

		for key, value := range node.Labels {
			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorMagenta),
				termenv.String(value).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Filter executors for this node
	var nodeExecutors []*core.Executor
	for _, executor := range executors {
		if executor.NodeID == node.ID {
			nodeExecutors = append(nodeExecutors, executor)
		}
	}

	// Print executors
	if len(nodeExecutors) > 0 {
		t, theme = createTable(0)
		t.SetTitle(fmt.Sprintf("Executors (%d)", len(nodeExecutors)))

		for i, executor := range nodeExecutors {
			row = []interface{}{
				termenv.String(fmt.Sprintf("#%d", i+1)).Foreground(theme.ColorBlue),
				termenv.String(fmt.Sprintf("%s (%s)", executor.Name, executor.ID)).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}
}
