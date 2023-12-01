package postgresql

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) {
	existingSnapshot, _ := db.GetSnapshotByName(colonyName, name)

	if existingSnapshot != nil {
		return nil, errors.New("Snapshot with name <" + name + "> in Colony <" + colonyName + "> already exists")
	}

	label = strings.TrimSuffix(label, "/")

	allLabels, err := db.GetFileLabelsByName(colonyName, label)
	if err != nil {
		return nil, err
	}

	snapshotID := core.GenerateRandomID()
	now := time.Now().UTC()

	var fileIDs []string
	for _, l := range allLabels {
		filenames, err := db.GetFilenamesByLabel(colonyName, l.Name)
		if err != nil {
			return nil, err
		}

		for _, filename := range filenames {
			file, err := db.GetLatestFileByName(colonyName, l.Name, filename)
			if err != nil {
				return nil, err
			}
			if len(file) != 1 {
				return nil, errors.New("failed to get file info, len>1")
			}
			fileIDs = append(fileIDs, file[0].ID)
		}

	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `SNAPSHOTS (SNAPSHOT_ID, COLONY_NAME, LABEL, NAME, FILE_IDS, ADDED) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.postgresql.Exec(sqlStatement, snapshotID, colonyName, label, colonyName+":"+name, pq.Array(fileIDs), now)
	if err != nil {
		return nil, err
	}
	snapshot := &core.Snapshot{ID: snapshotID, ColonyName: colonyName, Label: label, Name: name, FileIDs: fileIDs, Added: now}

	return snapshot, nil
}

func (db *PQDatabase) parseSnapshots(rows *sql.Rows) ([]*core.Snapshot, error) {
	var snapshots []*core.Snapshot

	for rows.Next() {
		var snapshotID string
		var colonyName string
		var label string
		var name string
		var fileIDs []string
		var added time.Time

		if err := rows.Scan(&snapshotID, &colonyName, &label, &name, pq.Array(&fileIDs), &added); err != nil {
			return nil, err
		}

		split := strings.Split(name, ":")

		if len(split) != 2 {
			return nil, errors.New("invalid split name")
		}

		snapshot := &core.Snapshot{ID: snapshotID, ColonyName: colonyName, Label: label, Name: split[1], FileIDs: fileIDs, Added: added}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (db *PQDatabase) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1 AND SNAPSHOT_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, snapshotID)
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
		return nil, errors.New("Snapshot not found with Id <" + snapshotID + "> in Colony <" + colonyName + "> does not exists")
	}
}

func (db *PQDatabase) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, colonyName+":"+name)
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
		return nil, errors.New("Snapshot not found with name <" + name + "> in Colony <" + colonyName + "> does not exists")
	}
}

func (db *PQDatabase) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1 ORDER BY ADDED DESC`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseSnapshots(rows)
}

func (db *PQDatabase) RemoveSnapshotByID(colonyName string, snapshotID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1 AND SNAPSHOT_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, snapshotID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveSnapshotByName(colonyName string, name string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1 AND NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, colonyName+":"+name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveSnapshotsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `SNAPSHOTS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}
