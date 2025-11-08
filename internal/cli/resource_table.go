package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

// printResourceDefinitionsTable displays a list of ResourceDefinitions in a table
func printResourceDefinitionsTable(rds []*core.ResourceDefinition) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "kind", Name: "Kind", SortIndex: 2},
		{ID: "executortype", Name: "ExecutorType", SortIndex: 3},
		{ID: "functionname", Name: "FunctionName", SortIndex: 4},
	}
	t.SetCols(cols)

	for _, rd := range rds {
		row := []interface{}{
			termenv.String(rd.Metadata.Name).Foreground(theme.ColorCyan),
			termenv.String(rd.Spec.Names.Kind).Foreground(theme.ColorViolet),
			termenv.String(rd.Spec.Handler.ExecutorType).Foreground(theme.ColorMagenta),
			termenv.String(rd.Spec.Handler.FunctionName).Foreground(theme.ColorBlue),
		}
		t.AddRow(row)
	}

	t.Render()
}

// printResourceDefinitionTable displays a single ResourceDefinition with details
func printResourceDefinitionTable(rd *core.ResourceDefinition) {
	t, theme := createTable(0)
	t.SetTitle("ResourceDefinition")

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(rd.Metadata.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ID").Foreground(theme.ColorCyan),
		termenv.String(rd.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Kind").Foreground(theme.ColorCyan),
		termenv.String(rd.Spec.Names.Kind).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Plural").Foreground(theme.ColorCyan),
		termenv.String(rd.Spec.Names.Plural).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Group").Foreground(theme.ColorCyan),
		termenv.String(rd.Spec.Group).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Version").Foreground(theme.ColorCyan),
		termenv.String(rd.Spec.Version).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Scope").Foreground(theme.ColorCyan),
		termenv.String(rd.Spec.Scope).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	// Handler section
	t, theme = createTable(0)
	t.SetTitle("Handler")

	row = []interface{}{
		termenv.String("Executor Type").Foreground(theme.ColorViolet),
		termenv.String(rd.Spec.Handler.ExecutorType).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Function Name").Foreground(theme.ColorViolet),
		termenv.String(rd.Spec.Handler.FunctionName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	if rd.Spec.Handler.ReconcileInterval > 0 {
		row = []interface{}{
			termenv.String("Reconcile Interval").Foreground(theme.ColorViolet),
			termenv.String(fmt.Sprintf("%d seconds", rd.Spec.Handler.ReconcileInterval)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	t.Render()

	// Schema section
	if rd.Spec.Schema != nil {
		t, theme = createTable(0)
		t.SetTitle("Schema")

		if len(rd.Spec.Schema.Required) > 0 {
			row = []interface{}{
				termenv.String("Required Fields").Foreground(theme.ColorBlue),
				termenv.String(strings.Join(rd.Spec.Schema.Required, ", ")).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		row = []interface{}{
			termenv.String("Properties").Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%d fields defined", len(rd.Spec.Schema.Properties))).Foreground(theme.ColorGray),
		}
		t.AddRow(row)

		t.Render()

		// Properties details
		if len(rd.Spec.Schema.Properties) > 0 {
			t, theme = createTable(0)
			t.SetTitle("Schema Properties")

			var propCols = []table.Column{
				{ID: "field", Name: "Field", SortIndex: 1},
				{ID: "type", Name: "Type", SortIndex: 2},
				{ID: "description", Name: "Description", SortIndex: 3},
			}
			t.SetCols(propCols)

			for propName, prop := range rd.Spec.Schema.Properties {
				desc := prop.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				if desc == "" {
					desc = "-"
				}

				row = []interface{}{
					termenv.String(propName).Foreground(theme.ColorMagenta),
					termenv.String(prop.Type).Foreground(theme.ColorCyan),
					termenv.String(desc).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			t.Render()
		}
	}
}

// printResourcesTable displays a list of Resources in a table
func printResourcesTable(resources []*core.Resource) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "kind", Name: "Kind", SortIndex: 2},
		{ID: "executortype", Name: "ExecutorType", SortIndex: 3},
		{ID: "functionname", Name: "FunctionName", SortIndex: 4},
		{ID: "generation", Name: "Gen", SortIndex: 5},
	}
	t.SetCols(cols)

	for _, resource := range resources {
		// Extract executorType and functionName from spec if they exist
		executorType := ""
		if val, ok := resource.GetSpec("executorType"); ok {
			if str, ok := val.(string); ok {
				executorType = str
			}
		}

		functionName := ""
		if val, ok := resource.GetSpec("functionName"); ok {
			if str, ok := val.(string); ok {
				functionName = str
			}
		}

		row := []interface{}{
			termenv.String(resource.Metadata.Name).Foreground(theme.ColorCyan),
			termenv.String(resource.Kind).Foreground(theme.ColorViolet),
			termenv.String(executorType).Foreground(theme.ColorMagenta),
			termenv.String(functionName).Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%d", resource.Metadata.Generation)).Foreground(theme.ColorYellow),
		}
		t.AddRow(row)
	}

	t.Render()
}

// printResourceTable displays a single Resource with details
func printResourceTable(client *client.ColoniesClient, resource *core.Resource) {
	t, theme := createTable(0)
	t.SetTitle("Resource")

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(resource.Metadata.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ID").Foreground(theme.ColorCyan),
		termenv.String(resource.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Kind").Foreground(theme.ColorCyan),
		termenv.String(resource.Kind).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Generation").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%d", resource.Metadata.Generation)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	// Reconciliation status
	if resource.Metadata.LastReconciliationProcess != "" {
		process, err := client.GetProcess(resource.Metadata.LastReconciliationProcess, PrvKey)
		if err == nil && process != nil {
			// Display reconciliation process ID
			row = []interface{}{
				termenv.String("Last Reconciliation").Foreground(theme.ColorCyan),
				termenv.String(process.ID).Foreground(theme.ColorGray),
			}
			t.AddRow(row)

			// Display reconciliation status with color coding
			statusColor := theme.ColorGray
			statusText := fmt.Sprintf("%d", process.State)
			switch process.State {
			case 0: // WAITING
				statusColor = theme.ColorYellow
				statusText = "WAITING"
			case 1: // RUNNING
				statusColor = theme.ColorCyan
				statusText = "RUNNING"
			case 2: // SUCCESS
				statusColor = theme.ColorGreen
				statusText = "SUCCESS"
			case 3: // FAILED
				statusColor = theme.ColorRed
				statusText = "FAILED"
			}

			row = []interface{}{
				termenv.String("Reconciliation Status").Foreground(theme.ColorCyan),
				termenv.String(statusText).Foreground(statusColor),
			}
			t.AddRow(row)

			// Display when reconciliation started
			if !resource.Metadata.LastReconciliationTime.IsZero() {
				row = []interface{}{
					termenv.String("Reconciliation Time").Foreground(theme.ColorCyan),
					termenv.String(resource.Metadata.LastReconciliationTime.Format(TimeLayout)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			// Display when process ended (if completed)
			if process.State == 2 || process.State == 3 { // SUCCESS or FAILED
				if !process.EndTime.IsZero() {
					row = []interface{}{
						termenv.String("Reconciliation Ended").Foreground(theme.ColorCyan),
						termenv.String(process.EndTime.Format(TimeLayout)).Foreground(theme.ColorGray),
					}
					t.AddRow(row)
				}
			}
		}
	}

	if !resource.Metadata.CreatedAt.IsZero() {
		row = []interface{}{
			termenv.String("Created At").Foreground(theme.ColorCyan),
			termenv.String(resource.Metadata.CreatedAt.Format(TimeLayout)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	if !resource.Metadata.UpdatedAt.IsZero() {
		row = []interface{}{
			termenv.String("Updated At").Foreground(theme.ColorCyan),
			termenv.String(resource.Metadata.UpdatedAt.Format(TimeLayout)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	t.Render()

	// Labels section
	if len(resource.Metadata.Labels) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Labels")

		for key, value := range resource.Metadata.Labels {
			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorViolet),
				termenv.String(value).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Annotations section
	if len(resource.Metadata.Annotations) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Annotations")

		for key, value := range resource.Metadata.Annotations {
			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorViolet),
				termenv.String(value).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Spec section
	if len(resource.Spec) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Spec")

		for key, value := range resource.Spec {
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 60 {
				valueStr = valueStr[:57] + "..."
			}

			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorMagenta),
				termenv.String(valueStr).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Status section
	if len(resource.Status) > 0 {
		// Check if this is a deployment status with containers
		if containers, ok := resource.Status["containers"].([]interface{}); ok && len(containers) > 0 {
			// Display deployment status summary
			t, theme = createTable(0)
			t.SetTitle("Deployment Status")

			if running, ok := resource.Status["runningReplicas"]; ok {
				row = []interface{}{
					termenv.String("Running Replicas").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", running)).Foreground(theme.ColorGreen),
				}
				t.AddRow(row)
			}

			if total, ok := resource.Status["totalReplicas"]; ok {
				row = []interface{}{
					termenv.String("Total Replicas").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", total)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			if lastUpdated, ok := resource.Status["lastUpdated"]; ok {
				row = []interface{}{
					termenv.String("Last Updated").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", lastUpdated)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			t.Render()

			// Display containers table
			t, theme = createTable(0)
			t.SetTitle("Containers")

			var containerCols = []table.Column{
				{ID: "name", Name: "Name", SortIndex: 1},
				{ID: "id", Name: "ID", SortIndex: 2},
				{ID: "state", Name: "State", SortIndex: 3},
				{ID: "image", Name: "Image", SortIndex: 4},
				{ID: "lastcheck", Name: "Last Check", SortIndex: 5},
			}
			t.SetCols(containerCols)

			for _, container := range containers {
				if containerMap, ok := container.(map[string]interface{}); ok {
					name := fmt.Sprintf("%v", containerMap["name"])
					id := fmt.Sprintf("%v", containerMap["id"])
					state := fmt.Sprintf("%v", containerMap["state"])
					image := fmt.Sprintf("%v", containerMap["image"])
					lastCheck := fmt.Sprintf("%v", containerMap["lastCheck"])

					// Parse and format lastCheck if it's a valid timestamp
					if lastCheck != "" && lastCheck != "<nil>" {
						if t, err := time.Parse(time.RFC3339, lastCheck); err == nil {
							lastCheck = t.Format("2006-01-02 15:04:05")
						}
					} else {
						lastCheck = "-"
					}

					// Color-code state
					stateColor := theme.ColorGray
					if state == "running" {
						stateColor = theme.ColorGreen
					} else if state == "stopped" {
						stateColor = theme.ColorRed
					}

					row = []interface{}{
						termenv.String(name).Foreground(theme.ColorCyan),
						termenv.String(id).Foreground(theme.ColorGray),
						termenv.String(state).Foreground(stateColor),
						termenv.String(image).Foreground(theme.ColorMagenta),
						termenv.String(lastCheck).Foreground(theme.ColorGray),
					}
					t.AddRow(row)
				}
			}

			t.Render()
		} else {
			// Generic status display for non-deployment resources
			t, theme = createTable(0)
			t.SetTitle("Status")

			for key, value := range resource.Status {
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 60 {
					valueStr = valueStr[:57] + "..."
				}

				row = []interface{}{
					termenv.String(key).Foreground(theme.ColorBlue),
					termenv.String(valueStr).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			t.Render()
		}
	}
}

// printResourceHistoryTable displays a list of ResourceHistory entries in a table
func printResourceHistoryTable(histories []*core.ResourceHistory) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "generation", Name: "Generation", SortIndex: 1},
		{ID: "timestamp", Name: "Timestamp", SortIndex: 2},
		{ID: "changetype", Name: "Change Type", SortIndex: 3},
		{ID: "changedby", Name: "Changed By", SortIndex: 4},
	}
	t.SetCols(cols)

	for _, history := range histories {
		row := []interface{}{
			termenv.String(fmt.Sprintf("%d", history.Generation)).Foreground(theme.ColorCyan),
			termenv.String(history.Timestamp.Format("2006-01-02 15:04:05")).Foreground(theme.ColorGray),
			termenv.String(history.ChangeType).Foreground(theme.ColorViolet),
			termenv.String(truncateString(history.ChangedBy, 40)).Foreground(theme.ColorBlue),
		}
		t.AddRow(row)
	}

	t.Render()
}

// truncateString truncates a string if it's longer than maxLen
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// printResourceHistoryDetail displays detailed information for a specific resource history entry
func printResourceHistoryDetail(history *core.ResourceHistory) {
	t, theme := createTable(0)
	t.SetTitle(fmt.Sprintf("Resource History - Generation %d", history.Generation))

	row := []interface{}{
		termenv.String("Resource ID").Foreground(theme.ColorCyan),
		termenv.String(history.ResourceID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Kind").Foreground(theme.ColorCyan),
		termenv.String(history.Kind).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(history.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Namespace").Foreground(theme.ColorCyan),
		termenv.String(history.Namespace).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Generation").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%d", history.Generation)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Timestamp").Foreground(theme.ColorCyan),
		termenv.String(history.Timestamp.Format("2006-01-02 15:04:05")).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Change Type").Foreground(theme.ColorCyan),
		termenv.String(history.ChangeType).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Changed By").Foreground(theme.ColorCyan),
		termenv.String(history.ChangedBy).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	// Spec section
	if len(history.Spec) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Spec")

		for key, value := range history.Spec {
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 60 {
				valueStr = valueStr[:57] + "..."
			}

			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorMagenta),
				termenv.String(valueStr).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Status section
	if len(history.Status) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Status")

		for key, value := range history.Status {
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 60 {
				valueStr = valueStr[:57] + "..."
			}

			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorBlue),
				termenv.String(valueStr).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}
}
