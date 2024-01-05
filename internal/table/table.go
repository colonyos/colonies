package table

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	width = 0 // max width
)

type TableOptions struct {
	Columns []int
	SortBy  int
	Style   table.Style
}

type Column struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

type Table struct {
	tab   table.Writer
	theme Theme
	opts  TableOptions
}

func NewTable(theme Theme, opts TableOptions, ascii bool) *Table {
	t := &Table{}
	t.tab = table.NewWriter()
	t.tab.SetAllowedRowLength(width)
	t.tab.SetOutputMirror(os.Stdout)
	t.tab.Style().Options.SeparateColumns = true

	if !ascii {
		t.tab.SetStyle(opts.Style)
	}

	t.theme = theme
	t.opts = opts

	return t
}

func (t *Table) SetTitle(title string) {
	t.tab.SetTitle("%s", title)
}

func (t *Table) SetCols(cols []Column) {
	headers := table.Row{}
	for _, v := range cols {
		headers = append(headers, v.Name)
	}
	t.tab.AppendHeader(headers)
}

func (t *Table) AddRow(row []interface{}) {
	t.tab.AppendRow(row)
}

func (t *Table) Render() {
	if t.tab.Length() == 0 {
		return
	}

	sortMode := table.Dsc
	if t.opts.SortBy >= 12 {
		sortMode = table.AscNumeric
	}

	t.tab.SortBy([]table.SortBy{{Number: t.opts.SortBy, Mode: sortMode}})
	t.tab.Render()
}
