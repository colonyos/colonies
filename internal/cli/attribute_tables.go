package cli

import (
	"fmt"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printAttributesTable(process *core.Process) {
	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t, theme := createTable(sortCol)

	t.SetTitle("Attributes")

	if len(process.Attributes) > 0 {
		var cols = []table.Column{
			{ID: "attributeid", Name: "AttributeId", SortIndex: 1},
			{ID: "key", Name: "Key", SortIndex: 2},
			{ID: "type", Name: "Type", SortIndex: 3},
		}
		t.SetCols(cols)

		for _, attribute := range process.Attributes {
			var attributeType string
			switch attribute.AttributeType {
			case core.IN:
				attributeType = "In"
			case core.OUT:
				attributeType = "Out"
			case core.ERR:
				attributeType = "Err"
			case core.ENV:
				attributeType = "Env"
			default:
				attributeType = "Unknown"
			}
			var key string
			if len(attribute.Key) > MaxAttributeLength {
				key = attribute.Key[0:MaxAttributeLength] + "..."
			} else {
				key = attribute.Key
			}

			var value string
			if len(attribute.Value) > MaxAttributeLength {
				value = attribute.Value[0:MaxAttributeLength] + "..."
			} else {
				value = attribute.Value
			}
			row := []interface{}{
				termenv.String(attribute.ID).Foreground(theme.ColorGray),
				termenv.String(key).Foreground(theme.ColorViolet),
				termenv.String(value).Foreground(theme.ColorCyan),
				termenv.String(attributeType).Foreground(theme.ColorMagenta),
			}
			t.AddRow(row)
		}
		t.Render()
	} else {
		fmt.Println("\nNo attributes found")
	}
}

func printAttributeTable(attribute *core.Attribute) {
	var sortCol int
	if ShowIDs {
		sortCol = 5
	} else {
		sortCol = 4
	}

	t, theme := createTable(sortCol)

	var cols = []table.Column{
		{ID: "attributeid", Name: "AttributeId", SortIndex: 1},
		{ID: "targetid", Name: "TargetId", SortIndex: 2},
		{ID: "key", Name: "Key", SortIndex: 3},
		{ID: "value", Name: "Value", SortIndex: 4},
		{ID: "type", Name: "Type", SortIndex: 5},
	}
	t.SetCols(cols)

	var attributeType string
	switch attribute.AttributeType {
	case core.IN:
		attributeType = "In"
	case core.OUT:
		attributeType = "Out"
	case core.ERR:
		attributeType = "Err"
	case core.ENV:
		attributeType = "Env"
	default:
		attributeType = "Unknown"
	}
	var key string
	if len(attribute.Key) > MaxAttributeLength {
		key = attribute.Key[0:MaxAttributeLength] + "..."
	} else {
		key = attribute.Key
	}

	var value string
	if len(attribute.Value) > MaxAttributeLength {
		value = attribute.Value[0:MaxAttributeLength] + "..."
	} else {
		value = attribute.Value
	}
	row := []interface{}{
		termenv.String(attribute.ID).Foreground(theme.ColorGray),
		termenv.String(attribute.TargetID).Foreground(theme.ColorCyan),
		termenv.String(key).Foreground(theme.ColorViolet),
		termenv.String(value).Foreground(theme.ColorViolet),
		termenv.String(attributeType).Foreground(theme.ColorMagenta),
	}
	t.AddRow(row)

	t.Render()
}
