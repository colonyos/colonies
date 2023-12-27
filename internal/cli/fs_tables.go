package cli

import (
	"errors"
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/muesli/termenv"
	log "github.com/sirupsen/logrus"
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

func printFilesTable(files []fileInfo) {
	sortCol := 1

	t, theme := createTable(sortCol)

	var cols = []table.Column{
		{ID: "filename", Name: "Filename", SortIndex: 1},
		{ID: "size", Name: "Size", SortIndex: 2},
		{ID: "latestid", Name: "Latest Id", SortIndex: 3},
		{ID: "added", Name: "Added", SortIndex: 4},
		{ID: "revisions", Name: "Revisions", SortIndex: 5},
	}
	t.SetCols(cols)

	for _, file := range files {
		row := []interface{}{
			termenv.String(file.filename).Foreground(theme.ColorCyan),
			termenv.String(file.size).Foreground(theme.ColorBlue),
			termenv.String(file.fileID).Foreground(theme.ColorMagenta),
			termenv.String(file.added.Format(TimeLayout)).Foreground(theme.ColorGreen),
			termenv.String(file.revisions).Foreground(theme.ColorYellow),
		}
		t.AddRow(row)
	}

	t.Render()
}

func printFileInfoTable(file *core.File) {
	sortCol := 0
	t, theme := createTable(sortCol)

	row := []interface{}{
		termenv.String("Filename").Foreground(theme.ColorCyan),
		termenv.String(file.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("FileId").Foreground(theme.ColorCyan),
		termenv.String(file.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Added").Foreground(theme.ColorCyan),
		termenv.String(file.Added.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Sequence Number").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatInt(file.SequenceNumber, 10)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Label").Foreground(theme.ColorCyan),
		termenv.String(file.Label).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(file.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Size").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatInt(file.Size/1024, 10) + " KiB").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Checksum").Foreground(theme.ColorCyan),
		termenv.String(file.Checksum).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Checksum Alg").Foreground(theme.ColorCyan),
		termenv.String(file.ChecksumAlg).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Protocol").Foreground(theme.ColorCyan),
		termenv.String(file.Reference.Protocol).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Endpoint").Foreground(theme.ColorCyan),
		termenv.String(file.Reference.S3Object.Server).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 TLS").Foreground(theme.ColorCyan),
		termenv.String(strconv.FormatBool(file.Reference.S3Object.TLS)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Region").Foreground(theme.ColorCyan),
		termenv.String(file.Reference.S3Object.Region).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Bucket").Foreground(theme.ColorCyan),
		termenv.String(file.Reference.S3Object.Bucket).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Object").Foreground(theme.ColorCyan),
		termenv.String(file.Reference.S3Object.Object).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Accesskey").Foreground(theme.ColorCyan),
		termenv.String("******************************").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("S3 Secretkey").Foreground(theme.ColorCyan),
		termenv.String("******************************").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Encryption Key").Foreground(theme.ColorCyan),
		termenv.String("******************************").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Encryption Alg").Foreground(theme.ColorCyan),
		termenv.String("").Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printSnapshotTable(snapshot *core.Snapshot, client *client.ColoniesClient) {
	sortCol := 0
	t, theme := createTable(sortCol)

	row := []interface{}{
		termenv.String("SnapshotId").Foreground(theme.ColorCyan),
		termenv.String(snapshot.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Name").Foreground(theme.ColorCyan),
		termenv.String(snapshot.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Label").Foreground(theme.ColorCyan),
		termenv.String(snapshot.Label).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(snapshot.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Added").Foreground(theme.ColorCyan),
		termenv.String(snapshot.Added.Format(TimeLayout)).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()

	sortCol = 0
	t, theme = createTable(sortCol)

	var cols = []table.Column{
		{ID: "filename", Name: "Filename", SortIndex: 1},
		{ID: "fileid", Name: "FileId", SortIndex: 2},
		{ID: "Added", Name: "Added", SortIndex: 3},
	}
	t.SetCols(cols)

	if len(snapshot.FileIDs) > 0 {
		for _, fileID := range snapshot.FileIDs {
			revision, err := client.GetFileByID(ColonyName, fileID, PrvKey)
			CheckError(err)
			if len(revision) != 1 {
				CheckError(errors.New("Expected only one revision"))
			}
			row = []interface{}{
				termenv.String(revision[0].Name).Foreground(theme.ColorViolet),
				termenv.String(fileID).Foreground(theme.ColorGray),
				termenv.String(revision[0].Added.Format(TimeLayout)).Foreground(theme.ColorBlue),
			}
			t.AddRow(row)
		}
		t.Render()
	} else {
		log.WithFields(log.Fields{"SnapshotID": SnapshotID, "SnapshotName": SnapshotName}).Warning("No files in snapshot")
	}
}

func printSnapshotsTable(snapshots []*core.Snapshot) {
	sortCol := 0
	t, theme := createTable(sortCol)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "snapshotid", Name: "SnapshotId", SortIndex: 2},
		{ID: "label", Name: "Label", SortIndex: 3},
		{ID: "Files", Name: "Files", SortIndex: 4},
		{ID: "Added", Name: "Added", SortIndex: 5},
	}
	t.SetCols(cols)

	for _, snapshot := range snapshots {
		row := []interface{}{
			termenv.String(snapshot.Name).Foreground(theme.ColorCyan),
			termenv.String(snapshot.ID).Foreground(theme.ColorGray),
			termenv.String(snapshot.Label).Foreground(theme.ColorViolet),
			termenv.String(strconv.Itoa(len(snapshot.FileIDs))).Foreground(theme.ColorMagenta),
			termenv.String(snapshot.Added.Format(TimeLayout)).Foreground(theme.ColorBlue),
		}
		t.AddRow(row)
	}

	t.Render()
}
