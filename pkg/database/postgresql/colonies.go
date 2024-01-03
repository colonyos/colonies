package postgresql

import (
	"database/sql"
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddColony(colony *core.Colony) error {
	if colony == nil {
		return errors.New("Colony is nil")
	}

	exitingColony, err := db.GetColonyByName(colony.Name)
	if err != nil {
		return err
	}

	if exitingColony != nil {
		return errors.New("Colony with name <" + colony.Name + "> already exists")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `COLONIES (COLONY_ID, NAME) VALUES ($1, $2)`
	_, err = db.postgresql.Exec(sqlStatement, colony.ID, colony.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseColonies(rows *sql.Rows) ([]*core.Colony, error) {
	var colonies []*core.Colony

	for rows.Next() {
		var name string
		var colonyID string
		if err := rows.Scan(&name, &colonyID); err != nil {
			return nil, err
		}

		colony := core.CreateColony(colonyID, name)
		colonies = append(colonies, colony)
	}

	return colonies, nil
}

func (db *PQDatabase) GetColonies() ([]*core.Colony, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COLONIES`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseColonies(rows)
}

func (db *PQDatabase) GetColonyByID(id string) (*core.Colony, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COLONIES WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	colonies, err := db.parseColonies(rows)
	if err != nil {
		return nil, err
	}

	if len(colonies) == 0 {
		return nil, nil
	}

	return colonies[0], nil
}

func (db *PQDatabase) GetColonyByName(name string) (*core.Colony, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COLONIES WHERE NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	colonies, err := db.parseColonies(rows)
	if err != nil {
		return nil, err
	}

	if len(colonies) == 0 {
		return nil, nil
	}

	return colonies[0], nil
}

func (db *PQDatabase) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error {
	sqlStatement := `UPDATE  ` + db.dbPrefix + `COLONIES SET COLONY_ID=$1 WHERE NAME=$2 AND COLONY_ID=$3`
	_, err := db.postgresql.Query(sqlStatement, newColonyID, colonyName, oldColonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RenameColony(colonyName string, newName string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `COLONIES SET NAME=$1 WHERE NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, newName, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveColonyByName(colonyName string) error {
	colony, err := db.GetColonyByName(colonyName)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony does not exists")
	}

	err = db.RemoveUsersByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveExecutorsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `COLONIES WHERE NAME=$1`
	_, err = db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	err = db.RemoveAllProcessesByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveAllProcessGraphsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveAllGeneratorsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveAllCronsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveFunctionsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveLogsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveFilesByColonyName(colony.Name)
	if err != nil {
		return err
	}

	err = db.RemoveSnapshotsByColonyName(colony.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountColonies() (int, error) {
	colonies, err := db.GetColonies()
	if err != nil {
		return -1, err
	}

	return len(colonies), nil
}
