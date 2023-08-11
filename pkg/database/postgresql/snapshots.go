package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) CreateSnapshot(colonyID string, label string, name string) (*core.Snapshot, error) {
	filenames, err := db.GetFilenamesByLabel(colonyID, label)
	if err != nil {
		return nil, err
	}
	var fileIDs []string

	for _, filename := range filenames {
		file, err := db.GetLatestFileByName(colonyID, label, filename)
		if err != nil {
			return nil, err
		}
		if len(file) != 1 {
			return nil, errors.New("failed to get file info, len>1")
		}
		fileIDs = append(fileIDs, file[0].ID)
	}

	snapshotID := core.GenerateRandomID()
	now := time.Now().UTC()
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `SNAPSHOTS (SNAPSHOT_ID, COLONY_ID, LABEL, NAME, FILE_IDS, ADDED) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.postgresql.Exec(sqlStatement, snapshotID, colonyID, label, name, pq.Array(fileIDs), now)
	if err != nil {
		return nil, err
	}

	snapshot := &core.Snapshot{ID: snapshotID, ColonyID: colonyID, Label: label, Name: name, FileIDs: fileIDs, Added: now}

	return snapshot, nil
}

func (db *PQDatabase) parseSnapshots(rows *sql.Rows) ([]*core.Snapshot, error) {
	var snapshots []*core.Snapshot

	for rows.Next() {
		var snapshotID string
		var colonyID string
		var label string
		var name string
		var fileIDs []string
		var added time.Time

		if err := rows.Scan(&snapshotID, &colonyID, &label, &name, pq.Array(&fileIDs), &added); err != nil {
			return nil, err
		}

		snapshot := &core.Snapshot{ID: snapshotID, ColonyID: colonyID, Label: label, Name: name, FileIDs: fileIDs, Added: added}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (db *PQDatabase) GetSnapshotByID(colonyID string, snapshotID string) (*core.Snapshot, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_ID=$1 AND SNAPSHOT_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, snapshotID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snapshots, err := db.parseSnapshots(rows)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 1 {
		return snapshots[0], nil
	} else {
		return nil, errors.New("Snapshot not found")
	}
}

func (db *PQDatabase) GetSnapshotsByColonyID(colonyID string) ([]*core.Snapshot, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_ID=$1 ORDER BY ADDED DESC`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseSnapshots(rows)
}

func (db *PQDatabase) DeleteSnapshotByID(colonyID string, snapshotID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_ID=$1 AND SNAPSHOT_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, snapshotID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteSnapshotsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
