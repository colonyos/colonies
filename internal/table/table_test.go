package table

import (
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	style := table.StyleRounded
	sortCol := 1
	theme, err := LoadTheme("dark")
	assert.Nil(t, err)
	table := NewTable(theme, TableOptions{
		Columns: []int{1, 2},
		SortBy:  sortCol,
		Style:   style,
	}, true)

	var cols = []Column{
		{ID: "mountpoint", Name: "Mounted on", SortIndex: 1},
		{ID: "filesystem", Name: "Filesystem", SortIndex: 2},
	}

	table.SetCols(cols)

	row := []interface{}{
		termenv.String("hej").Foreground(theme.ColorBlue),  // mounted on
		termenv.String("test").Foreground(theme.ColorGray), // filesystem
	}

	table.AddRow(row)

	table.Render()
}
