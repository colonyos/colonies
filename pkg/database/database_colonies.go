package database

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

func (db *Database) AddColony(colony *core.Colony) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `COLONIES (COLONY_ID, PRIVATE_KEY, NAME) VALUES ($1, $2, $3)`
	_, err := db.postgresql.Exec(sqlStatement, colony.ID(), colony.PrivateKey(), colony.Name())
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) parseColonies(rows *sql.Rows) ([]*core.Colony, error) {
	var colonies []*core.Colony

	for rows.Next() {
		var id string
		var privateKey string
		var name string
		if err := rows.Scan(&id, &privateKey, &name); err != nil {
			return nil, err
		}

		colony, err := core.CreateColonyFromDB(name, privateKey)
		if err != nil {
			return nil, err
		}
		colonies = append(colonies, colony)
	}

	return colonies, nil
}

func (db *Database) GetColonies() ([]*core.Colony, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `COLONIES`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseColonies(rows)
}

func (db *Database) GetColonyByID(id string) (*core.Colony, error) {
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

	if len(colonies) > 1 {
		return nil, errors.New("expected one colony, colony id should be unique")
	}

	if len(colonies) == 0 {
		return nil, nil
	}

	return colonies[0], nil
}

func (db *Database) DeleteColonyByID(colonyID string) error {
	err := db.DeleteWorkersByColonyID(colonyID)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `COLONIES WHERE COLONY_ID=$1`
	_, err = db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
