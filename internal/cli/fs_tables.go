package cli

import (
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/muesli/termenv"
)

func printSyncPlans(syncPlans []*fs.SyncPlan) {
	remoteMissingTable, theme := createTable(3)
	remoteMissingTable.SetTitle("These files will be uploaded")
	var cols = []table.Column{
		{ID: "file", Name: "File", SortIndex: 1},
		{ID: "size", Name: "Size", SortIndex: 2},
		{ID: "label", Name: "Label", SortIndex: 3},
	}
	remoteMissingTable.SetCols(cols)

	localMissingTable, theme := createTable(3)
	localMissingTable.SetTitle("These files will be downloaded to " + SyncDir)
	localMissingTable.SetCols(cols)

	conflictTable, theme := createTable(3)
	if KeepLocal {
		conflictTable.SetTitle("These files will be replaced at the server")
	} else {
		conflictTable.SetTitle("These files will be replaced locally")
	}
	conflictTable.SetCols(cols)

	for _, syncPlan := range syncPlans {
		if len(syncPlan.RemoteMissing) > 0 {
			for _, file := range syncPlan.RemoteMissing {
				row := []interface{}{
					termenv.String(file.Name).Foreground(theme.ColorBlue),
					termenv.String(strconv.FormatInt(file.Size/1024, 10) + " KiB").Foreground(theme.ColorCyan),
					termenv.String(syncPlan.Label).Foreground(theme.ColorViolet),
				}
				remoteMissingTable.AddRow(row)
			}
		}

		if len(syncPlan.LocalMissing) > 0 {
			for _, file := range syncPlan.LocalMissing {
				row := []interface{}{
					termenv.String(file.Name).Foreground(theme.ColorBlue),
					termenv.String(strconv.FormatInt(file.Size/1024, 10) + " KiB").Foreground(theme.ColorCyan),
					termenv.String(syncPlan.Label).Foreground(theme.ColorViolet),
				}
				localMissingTable.AddRow(row)
			}
		}

		if len(syncPlan.Conflicts) > 0 {
			for _, file := range syncPlan.Conflicts {
				row := []interface{}{
					termenv.String(file.Name).Foreground(theme.ColorBlue),
					termenv.String(strconv.FormatInt(file.Size/1024, 10) + " KiB").Foreground(theme.ColorCyan),
					termenv.String(syncPlan.Label).Foreground(theme.ColorViolet),
				}
				conflictTable.AddRow(row)
			}
		}
	}
	remoteMissingTable.Render()
	localMissingTable.Render()
	conflictTable.Render()
}

func printLabelsTable(labels []*core.Label) {
	sortCol := 1

	t, theme := createTable(sortCol)

	var cols = []table.Column{
		{ID: "label", Name: "Label", SortIndex: 1},
		{ID: "files", Name: "Files", SortIndex: 2},
	}
	t.SetCols(cols)

	for _, label := range labels {
		row := []interface{}{
			termenv.String(label.Name).Foreground(theme.ColorCyan),
			termenv.String(strconv.Itoa(label.Files)).Foreground(theme.ColorBlue),
		}
		t.AddRow(row)
	}

	t.Render()
}
