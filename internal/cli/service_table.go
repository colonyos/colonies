package cli

import (
	"fmt"
	"sort"
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

// printResourcesTable displays a list of Services in a table
func printResourcesTable(services []*core.Service) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "kind", Name: "Kind", SortIndex: 2},
		{ID: "executortype", Name: "ExecutorType", SortIndex: 3},
		{ID: "functionname", Name: "FunctionName", SortIndex: 4},
		{ID: "generation", Name: "Gen", SortIndex: 5},
	}
	t.SetCols(cols)

	for _, service := range services {
		// Extract executorType and functionName from spec if they exist
		executorType := ""
		if val, ok := service.GetSpec("executorType"); ok {
			if str, ok := val.(string); ok {
				executorType = str
			}
		}

		functionName := ""
		if val, ok := service.GetSpec("functionName"); ok {
			if str, ok := val.(string); ok {
				functionName = str
			}
		}

		row := []interface{}{
			termenv.String(service.Metadata.Name).Foreground(theme.ColorCyan),
			termenv.String(service.Kind).Foreground(theme.ColorViolet),
			termenv.String(executorType).Foreground(theme.ColorMagenta),
			termenv.String(functionName).Foreground(theme.ColorBlue),
			termenv.String(fmt.Sprintf("%d", service.Metadata.Generation)).Foreground(theme.ColorYellow),
		}
		t.AddRow(row)
	}

	t.Render()
}

// printResourceTable displays a single Service with details
func printResourceTable(client *client.ColoniesClient, service *core.Service) {
	t, theme := createTable(0)
	t.SetTitle("Service")

	row := []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(service.Metadata.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("ID").Foreground(theme.ColorCyan),
		termenv.String(service.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Kind").Foreground(theme.ColorCyan),
		termenv.String(service.Kind).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Generation").Foreground(theme.ColorCyan),
		termenv.String(fmt.Sprintf("%d", service.Metadata.Generation)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	// Reconciliation status
	if service.Metadata.LastReconciliationProcess != "" {
		process, err := client.GetProcess(service.Metadata.LastReconciliationProcess, PrvKey)
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
			if !service.Metadata.LastReconciliationTime.IsZero() {
				row = []interface{}{
					termenv.String("Reconciliation Time").Foreground(theme.ColorCyan),
					termenv.String(service.Metadata.LastReconciliationTime.Format(TimeLayout)).Foreground(theme.ColorGray),
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

	if !service.Metadata.CreatedAt.IsZero() {
		row = []interface{}{
			termenv.String("Created At").Foreground(theme.ColorCyan),
			termenv.String(service.Metadata.CreatedAt.Format(TimeLayout)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	if !service.Metadata.UpdatedAt.IsZero() {
		row = []interface{}{
			termenv.String("Updated At").Foreground(theme.ColorCyan),
			termenv.String(service.Metadata.UpdatedAt.Format(TimeLayout)).Foreground(theme.ColorGray),
		}
		t.AddRow(row)
	}

	t.Render()

	// Labels section
	if len(service.Metadata.Labels) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Labels")

		for key, value := range service.Metadata.Labels {
			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorViolet),
				termenv.String(value).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Annotations section
	if len(service.Metadata.Annotations) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Annotations")

		for key, value := range service.Metadata.Annotations {
			row = []interface{}{
				termenv.String(key).Foreground(theme.ColorViolet),
				termenv.String(value).Foreground(theme.ColorGray),
			}
			t.AddRow(row)
		}

		t.Render()
	}

	// Spec section
	if len(service.Spec) > 0 {
		t, theme = createTable(0)
		t.SetTitle("Spec")

		// Define special keys that need formatted display
		complexKeys := map[string]bool{
			"env":     true,
			"volumes": true,
			"ports":   true,
		}

		for key, value := range service.Spec {
			// Skip complex fields - they'll be rendered separately
			if complexKeys[key] {
				continue
			}

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

		// Render environment variables if present
		if env, ok := service.Spec["env"].(map[string]interface{}); ok && len(env) > 0 {
			t, theme = createTable(0)
			t.SetTitle("Environment Variables")

			var envCols = []table.Column{
				{ID: "key", Name: "Key", SortIndex: 1},
				{ID: "value", Name: "Value", SortIndex: 2},
			}
			t.SetCols(envCols)

			// Sort keys alphabetically
			keys := make([]string, 0, len(env))
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				valueStr := fmt.Sprintf("%v", env[k])
				row = []interface{}{
					termenv.String(k).Foreground(theme.ColorCyan),
					termenv.String(valueStr).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			t.Render()
		}

		// Render volumes if present
		if volumes, ok := service.Spec["volumes"].([]interface{}); ok && len(volumes) > 0 {
			t, theme = createTable(0)
			t.SetTitle("Volumes")

			var volCols = []table.Column{
				{ID: "host", Name: "Host Path", SortIndex: 1},
				{ID: "container", Name: "Container Path", SortIndex: 2},
				{ID: "readonly", Name: "Read Only", SortIndex: 3},
			}
			t.SetCols(volCols)

			for _, vol := range volumes {
				if volMap, ok := vol.(map[string]interface{}); ok {
					host := fmt.Sprintf("%v", volMap["host"])
					container := fmt.Sprintf("%v", volMap["container"])
					readOnly := "false"
					if ro, ok := volMap["readOnly"].(bool); ok && ro {
						readOnly = "true"
					}

					row = []interface{}{
						termenv.String(host).Foreground(theme.ColorCyan),
						termenv.String(container).Foreground(theme.ColorMagenta),
						termenv.String(readOnly).Foreground(theme.ColorGray),
					}
					t.AddRow(row)
				}
			}

			t.Render()
		}

		// Render ports if present
		if ports, ok := service.Spec["ports"].([]interface{}); ok && len(ports) > 0 {
			t, theme = createTable(0)
			t.SetTitle("Ports")

			var portCols = []table.Column{
				{ID: "name", Name: "Name", SortIndex: 1},
				{ID: "port", Name: "Port", SortIndex: 2},
				{ID: "protocol", Name: "Protocol", SortIndex: 3},
			}
			t.SetCols(portCols)

			for _, port := range ports {
				if portMap, ok := port.(map[string]interface{}); ok {
					name := fmt.Sprintf("%v", portMap["name"])
					portNum := fmt.Sprintf("%v", portMap["port"])
					protocol := fmt.Sprintf("%v", portMap["protocol"])

					row = []interface{}{
						termenv.String(name).Foreground(theme.ColorCyan),
						termenv.String(portNum).Foreground(theme.ColorYellow),
						termenv.String(protocol).Foreground(theme.ColorGray),
					}
					t.AddRow(row)
				}
			}

			t.Render()
		}
	}

	// Status section
	if len(service.Status) > 0 {
		// Check if this is a deployment status with instances
		if instances, ok := service.Status["instances"].([]interface{}); ok && len(instances) > 0 {
			// Display deployment status summary
			t, theme = createTable(0)
			t.SetTitle("Deployment Status")

			if running, ok := service.Status["runningInstances"]; ok {
				row = []interface{}{
					termenv.String("Running Instances").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", running)).Foreground(theme.ColorGreen),
				}
				t.AddRow(row)
			}

			if total, ok := service.Status["totalInstances"]; ok {
				row = []interface{}{
					termenv.String("Total Instances").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", total)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			if lastUpdated, ok := service.Status["lastUpdated"]; ok {
				row = []interface{}{
					termenv.String("Last Updated").Foreground(theme.ColorBlue),
					termenv.String(fmt.Sprintf("%v", lastUpdated)).Foreground(theme.ColorGray),
				}
				t.AddRow(row)
			}

			t.Render()

			// Display instances table
			t, theme = createTable(0)
			t.SetTitle("Instances")

			var instanceCols = []table.Column{
				{ID: "name", Name: "Name", SortIndex: 1},
				{ID: "id", Name: "ID", SortIndex: 2},
				{ID: "type", Name: "Type", SortIndex: 3},
				{ID: "state", Name: "State", SortIndex: 4},
				{ID: "image", Name: "Image", SortIndex: 5},
				{ID: "lastcheck", Name: "Last Check", SortIndex: 6},
			}
			t.SetCols(instanceCols)

			for _, instance := range instances {
				if instanceMap, ok := instance.(map[string]interface{}); ok {
					name := fmt.Sprintf("%v", instanceMap["name"])
					id := fmt.Sprintf("%v", instanceMap["id"])
					instanceType := fmt.Sprintf("%v", instanceMap["type"])
					state := fmt.Sprintf("%v", instanceMap["state"])
					image := fmt.Sprintf("%v", instanceMap["image"])
					lastCheck := fmt.Sprintf("%v", instanceMap["lastCheck"])

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
						termenv.String(instanceType).Foreground(theme.ColorYellow),
						termenv.String(state).Foreground(stateColor),
						termenv.String(image).Foreground(theme.ColorMagenta),
						termenv.String(lastCheck).Foreground(theme.ColorGray),
					}
					t.AddRow(row)
				}
			}

			t.Render()
		} else {
			// Generic status display for non-deployment services
			t, theme = createTable(0)
			t.SetTitle("Status")

			// Filter out instance-related fields that should only show in the instances table
			excludeKeys := map[string]bool{
				"instances":        true,
				"runningInstances": true,
				"stoppedInstances": true,
				"totalInstances":   true,
			}

			for key, value := range service.Status {
				// Skip instance-related fields
				if excludeKeys[key] {
					continue
				}

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

			// Only render if there are actually rows to display
			if len(service.Status) > len(excludeKeys) {
				t.Render()
			}
		}
	}
}

// printResourceHistoryTable displays a list of ResourceHistory entries in a table
func printResourceHistoryTable(c *client.ColoniesClient, histories []*core.ResourceHistory) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "generation", Name: "Generation", SortIndex: 1},
		{ID: "timestamp", Name: "Timestamp", SortIndex: 2},
		{ID: "changetype", Name: "Change Type", SortIndex: 3},
		{ID: "changedby", Name: "Changed By", SortIndex: 4},
	}
	t.SetCols(cols)

	for _, history := range histories {
		changedByStr := truncateString(history.ChangedBy, 12)

		// Try to resolve ChangedBy ID to executor or user name
		executor, err := c.GetExecutorByID(ColonyName, history.ChangedBy, PrvKey)
		if err == nil && executor != nil {
			changedByStr = fmt.Sprintf("executor: %s", executor.Name)
		} else {
			// Try user lookup
			user, err := c.GetUserByID(ColonyName, history.ChangedBy, PrvKey)
			if err == nil && user != nil {
				changedByStr = fmt.Sprintf("user: %s", user.Name)
			}
		}

		row := []interface{}{
			termenv.String(fmt.Sprintf("%d", history.Generation)).Foreground(theme.ColorCyan),
			termenv.String(history.Timestamp.Format("2006-01-02 15:04:05")).Foreground(theme.ColorGray),
			termenv.String(history.ChangeType).Foreground(theme.ColorViolet),
			termenv.String(changedByStr).Foreground(theme.ColorBlue),
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

// printResourceHistoryDetail displays detailed information for a specific service history entry
func printResourceHistoryDetail(history *core.ResourceHistory) {
	t, theme := createTable(0)
	t.SetTitle(fmt.Sprintf("Service History - Generation %d", history.Generation))

	row := []interface{}{
		termenv.String("Service ID").Foreground(theme.ColorCyan),
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
