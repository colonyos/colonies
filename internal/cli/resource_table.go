package cli

import (
	"fmt"
	"strings"

	"github.com/colonyos/colonies/internal/table"
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
func printResourceTable(resource *core.Resource) {
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
